package hls

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"path"
	"strings"
	"sync"
)

type M3u8Parser interface {
	Parse(content, m3u8Url string) (string, error)
}

func NewM3u8Parser(ctx context.Context, wg *sync.WaitGroup, outputDir, outputFile string) M3u8Parser {
	return &m3u8Parser{
		outputDir:      outputDir,
		writer:         NewM3u8Writer(ctx, wg, outputDir, outputFile),
		itemDownloader: NewItemDownloader(ctx, wg, outputDir),
	}
}

type m3u8Parser struct {
	outputDir      string
	writer         M3u8Writer
	itemDownloader ItemDownloader
}

func (p *m3u8Parser) Parse(content, m3u8Url string) (string, error) {
	// 边解析边写入m3u8Writer，并构造item连接发到下载channel
	lines := strings.Split(content, "\n")

	// 检查是否是多码率自适应
	isAdaptive := false
	for _, line := range lines {
		if strings.HasPrefix(line, "#EXT-X-STREAM-INF") {
			isAdaptive = true
		} else if isAdaptive {
			return line, nil
		}
	}

	// 找出当前content中已写入item的index
	lastItemNameIdx := -1
	if lastItemName, err := p.writer.GetLastItemName(); err != nil {
		// first m3u8 request, do nothing
	} else {
		for i, line := range lines {
			if line == lastItemName {
				lastItemNameIdx = i
				break
			}
		}
	}

	// 开始解析
	gotItem := false
	var itemInfo *item
	itemIdx := -1
	for len(lines) != 0 {
		curLine := lines[0]
		if len(curLine) == 0 {
			lines = lines[1:]
			continue
		}
		if strings.HasPrefix(curLine, "#") {
			if strings.Contains(curLine, ":") {
				// tag:value
				tagvalue := strings.Split(curLine, ":")
				tag := tagvalue[0]
				value := tagvalue[1]
				if !strings.HasPrefix(tag, "#EXTINF") {
					p.writer.WriteHeader(&tag, &value)
				} else {
					gotItem = true
					itemInfo = &item{}
					itemInfo.Extinf = value
				}
			} else {
				log.Printf("ignore line:%s\n", curLine)
			}
		} else {
			// item line
			if !gotItem {
				log.Fatalf(`item line "%s" without #EXTINF tag\n%s`, curLine, content)
				return "", fmt.Errorf("item line without #EXTINF tag")
			}
			gotItem = false // 重置gotItem
			itemIdx++
			itemInfo.Name = curLine
			if itemIdx > lastItemNameIdx {
				// 写入writer
				p.writer.WriteItem(itemInfo)
				// 通知downloader
				err := p.notifyDownloader(itemInfo.Name, m3u8Url)
				if err != nil {
					return "", err
				}
			} else {
				log.Printf("skip duplicate item:%s", curLine)
			}
		}
		lines = lines[1:]
	}

	return "", nil
}

func (p *m3u8Parser) notifyDownloader(name, m3u8Url string) error {
	url, err := url.Parse(m3u8Url)
	if err != nil {
		return err
	}
	log.Printf("old url: %s\n", url)
	m3u8Dir := path.Dir(url.Path)

	fileUrl, err := url.Parse(name)
	if err != nil {
		return err
	}
	fileName := path.Base(fileUrl.Path)
	//fileQuery := fileUrl.RawQuery

	log.Printf(`m3u8Dir:%s, fileName:%s`, m3u8Dir, fileName)
	url.Path = path.Join(m3u8Dir, fileName)
	url.RawQuery = ""
	log.Printf("new url: %s\n", url)
	p.itemDownloader.AppendItemInfo(url.String())
	return nil
}
