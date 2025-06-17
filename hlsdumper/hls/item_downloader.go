package hls

import (
	"context"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"sync"
	"time"
)

type ItemDownloader interface {
	AppendItemInfo(url string)
}

func NewItemDownloader(ctx context.Context, wg *sync.WaitGroup, outputDir string) ItemDownloader {
	downloader := &itemDownloader{
		ctx:       ctx,
		wg:        wg,
		outputDir: outputDir,
		queue:     make(chan *downloadInfo),
	}

	downloader.processDownloading()

	return downloader
}

type downloadInfo struct {
	url string
}

type itemDownloader struct {
	ctx       context.Context
	wg        *sync.WaitGroup
	outputDir string
	queue     chan *downloadInfo
}

func (i *itemDownloader) AppendItemInfo(url string) {
	i.queue <- &downloadInfo{
		url: url,
	}
}

func (i *itemDownloader) processDownloading() {
	i.wg.Add(1)
	go func() {
		defer i.wg.Done()
		for {
			select {
			case <-i.ctx.Done():
				log.Printf("downloader exit")
				return
			case itemInfo := <-i.queue:
				i.download(itemInfo)
			}
		}
	}()
}

func (i *itemDownloader) download(itemInfo *downloadInfo) {
	filename := path.Base(itemInfo.url)
	outputFilename := path.Join(i.outputDir, filename)

	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(itemInfo.url)
	if err != nil {
		log.Printf("download file failed:%v", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Printf("http error:%s", resp.Status)
		return
	}

	outFile, err := os.Create((outputFilename))
	if err != nil {
		log.Printf("create file failed:%v", err)
		return
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		log.Printf("save file failed:%v", err)
		return
	}
	log.Printf("download %s to %s ok", itemInfo.url, outputFilename)
}
