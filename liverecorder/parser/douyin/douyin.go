package douyin

import (
	"context"
	"fmt"
	"io"
	"liverecorder/utils"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/tidwall/gjson"
)

const (
	referer            = "https://www.douyin.com/"
	randomCookieChars  = "1234567890abcdef"
	roomStatusRegex    = `self.__pace_f.push\(\[1,\s*"[^:]*:([^<]*,null,\{\\"state\\"[^<]*\])\\n"\]\)`
	streamUrlInfoRegex = `self.__pace_f.push\(\[1,\s*\"(\{.*?\})\"\]\)`
)

func createRandomCookie() string {
	return utils.GenRandomString(21, randomCookieChars)
}

type LiveRoom struct {
	LiveUrl  string
	streamId string
}

func (d *LiveRoom) GetStreamUrl() (string, error) {
	body, err := d.getLiveWebPage()
	if err != nil {
		return "", err
	}
	log.Printf("fetch body: %s\n", body)

	streamUrlInfos, err := d.getStreamUrlInfoFromBody(body)
	if err != nil {
		return "", err
	}
	if len(streamUrlInfos) == 0 {
		return "", fmt.Errorf("no stream url found")
	}

	streamUrl := streamUrlInfos[0].Url.String()
	for _, streamUrlInfo := range streamUrlInfos {
		//log.Printf("streamUrlInfo: %+v", streamUrlInfo)
		if strings.Contains(streamUrlInfo.Url.String(), "or4") {
			streamUrl = streamUrlInfo.Url.String()
		}
	}
	log.Printf(`found %d stream urls`, len(streamUrlInfos))
	return streamUrl, nil
}

func (d *LiveRoom) getStreamUrlInfoFromBody(body string) ([]*utils.StreamUrlInfo, error) {
	if isLiving, err := d.isLiving(body); err != nil {
		return nil, err
	} else if !isLiving {
		return nil, fmt.Errorf("not living")
	}
	// parse all stream urls
	streamUrlInfos := make([]*utils.StreamUrlInfo, 0)
	reg, err := regexp.Compile(streamUrlInfoRegex)
	if err != nil {
		return nil, err
	}
	match := reg.FindAllStringSubmatch(body, -1)
	if match == nil {
		return nil, fmt.Errorf("no match found:%s", streamUrlInfoRegex)
	}
	for _, item := range match {
		if len(item) < 2 {
			log.Printf(`invalid item: %v`, item)
			continue
		}
		streamUrlInfoLine := item[1]
		streamUrlInfoJson := gjson.Parse(gjson.Parse(fmt.Sprintf(`"%s"`, streamUrlInfoLine)).String())
		if !streamUrlInfoJson.Exists() {
			log.Printf(`invalid streamUrlInfoJson: %s`, streamUrlInfoLine)
			continue
		}

		streamId := streamUrlInfoJson.Get("common.stream").String()
		if streamId == "" {
			log.Printf("no local streamId found")
			continue
		}
		if streamId != d.streamId {
			log.Printf("streamId mismatch, local: %s, global: %s", streamId, d.streamId)
			continue
		}
		infos := d.parseStreamUrlData(streamUrlInfoJson.Get("data"))
		streamUrlInfos = append(streamUrlInfos, infos...)
	}
	return streamUrlInfos, nil
}

func (d *LiveRoom) parseStreamUrlData(data gjson.Result) []*utils.StreamUrlInfo {
	streamUrlInfos := make([]*utils.StreamUrlInfo, 0)
	data.ForEach(func(key, value gjson.Result) bool {
		flv := value.Get("main.flv").String()
		url, err := url.Parse(flv)
		if err != nil {
			log.Printf("invalid url: %s", flv)
			return true
		}
		paramString := value.Get("main.sdk_params").String()
		paramJson := gjson.Parse(paramString)
		var description strings.Builder
		paramJson.ForEach(func(key, value gjson.Result) bool {
			description.WriteString(key.String())
			description.WriteString(":")
			description.WriteString(value.String())
			description.WriteString("\n")
			return true
		})
		resolution := 0
		resolutionStrs := strings.Split(paramJson.Get("resolution").String(), "x")
		if len(resolutionStrs) == 2 {
			x, err := strconv.Atoi(resolutionStrs[0])
			if err != nil {
				log.Printf("invalid resolution: %s", paramJson.Get("resolution").String())
				return true
			}
			y, err := strconv.Atoi(resolutionStrs[1])
			if err != nil {
				log.Printf("invalid resolution: %s", paramJson.Get("resolution").String())
				return true
			}
			resolution = x * y
		}
		vbitrate := int(paramJson.Get("vbitrate").Int())

		streamUrlInfos = append(streamUrlInfos, &utils.StreamUrlInfo{
			Name:        key.String(),
			Description: description.String(),
			Url:         url,
			Resolution:  resolution,
			Vbitrate:    vbitrate,
		})
		return true
	})
	return streamUrlInfos
}

func (d *LiveRoom) isLiving(body string) (bool, error) {
	roomStatusJson, err := d.getRoomStatusJson(body)
	if err != nil {
		return false, err
	}

	// 记录streamId后续检查
	d.streamId = roomStatusJson.Get("state.streamStore.streamData.H264_streamData.common.stream").String()
	if d.streamId == "" {
		return false, fmt.Errorf(`no streamId found`)
	}

	isLiving := roomStatusJson.Get("state.roomStore.roomInfo.room.status_str").String() == "2"
	return isLiving, nil
}

func (d *LiveRoom) getRoomStatusJson(body string) (*gjson.Result, error) {
	reg, err := regexp.Compile(roomStatusRegex)
	if err != nil {
		return nil, err
	}
	match := reg.FindAllStringSubmatch(body, -1)
	if match == nil {
		return nil, fmt.Errorf("no match found:%s", roomStatusRegex)
	}

	for _, item := range match {
		if len(item) < 2 {
			log.Printf(`invalid item: %v`, item)
			continue
		}
		roomStatusLine := item[1]
		log.Printf("roomStatusLine: %s", roomStatusLine)

		roomStatusJson := gjson.Parse(fmt.Sprintf(`"%s"`, roomStatusLine))
		if !roomStatusJson.Exists() {
			log.Printf(`invalid roomStatusJson: %s`, roomStatusLine)
			continue
		}
		log.Printf("before get3 roomStatusJson:%s", roomStatusJson.String())
		roomStatusJson = gjson.Parse(roomStatusJson.String()).Get("3")
		if !roomStatusJson.Exists() {
			log.Printf(`invalid roomStatusJson: %s`, roomStatusJson.String())
			continue
		}
		log.Printf("after get3 roomStatusJson:%s", roomStatusJson.String())
		if roomStatusJson.Get("state.roomStore.roomInfo.room.status_str").Exists() {
			return &roomStatusJson, nil
		}
	}
	return nil, fmt.Errorf("no roomStatusJson found")
}

func (d *LiveRoom) getLiveWebPage() (string, error) {
	req, err := http.NewRequestWithContext(context.Background(), "GET", d.LiveUrl, nil)
	if err != nil {
		return "", err
	}
	// set headers
	req.Header.Set("User-Agent", utils.UserAgent)
	req.Header.Set("Referer", referer)

	// set queries
	queries := req.URL.Query()
	// add queries here
	req.URL.RawQuery = queries.Encode()

	// set cookies
	req.AddCookie(&http.Cookie{Name: "__ac_nonce", Value: createRandomCookie()})

	// do request
	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// check response
	switch code := resp.StatusCode; code {
	case http.StatusOK:
		// do something
	default:
		return "", fmt.Errorf("failed to get live web page, status code: %d", code)
	}
	rspBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(rspBody), nil
}
