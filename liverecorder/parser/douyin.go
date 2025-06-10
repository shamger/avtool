package parser

import (
	"context"
	"fmt"
	"io"
	"liverecorder/utils"
	"log"
	"net/http"
)

const (
	referer           = "https://www.douyin.com/"
	randomCookieChars = "1234567890abcdef"
)

func createRandomCookie() string {
	return utils.GenRandomString(21, randomCookieChars)
}

type douyin struct {
	liveUrl string
}

func (d *douyin) GetStreamUrl() (string, error) {
	body, err := d.getLiveWebPage()
	if err != nil {
		return "", err
	}
	log.Printf("fetch body: %s\n", body)
	return "", nil
}

func (d *douyin) getLiveWebPage() (string, error) {
	req, err := http.NewRequestWithContext(context.Background(), "GET", d.liveUrl, nil)
	if err != nil {
		return "", err
	}
	// set headers
	req.Header.Set("User-Agent", UserAgent)
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
