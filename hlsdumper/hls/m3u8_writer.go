package hls

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
)

type M3u8Writer interface {
	WriteHeader(tag, value *string)
	WriteItem(item *item)
	GetLastItemName() (string, error)
}

func NewM3u8Writer(ctx context.Context, wg *sync.WaitGroup, outputDir, outputFile string) M3u8Writer {
	writer := &m3u8Writer{
		ctx:        ctx,
		wg:         wg,
		outputDir:  outputDir,
		outputFile: outputFile,
		begin:      "#EXTM3U\n",
		header:     &header{},
		items:      []*item{},
		end:        "#EXT-X-ENDLIST\n",
	}
	writer.startLoop()
	return writer
}

type m3u8Writer struct {
	ctx        context.Context
	wg         *sync.WaitGroup
	outputDir  string
	outputFile string
	begin      string
	header     *header
	items      []*item
	end        string
}

func (w *m3u8Writer) startLoop() {
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		for {
			<-w.ctx.Done()
			log.Printf("m3u8 writer done\n")
			w.writerM3u8File()
			return
		}
	}()
}

func (w *m3u8Writer) writerM3u8File() {
	fileName := fmt.Sprintf("%s/%s", w.outputDir, w.outputFile)
	log.Printf("start to write m3u8 file: %s\n", fileName)
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("open file error: %v\n", err)
		return
	}
	defer file.Close()

	file.WriteString(w.begin)
	file.WriteString(fmt.Sprintf("#EXT-X-VERSION:%s\n", w.header.Version))
	file.WriteString(fmt.Sprintf("#EXT-X-MEDIA-SEQUENCE:%s\n", w.header.Sequence))
	file.WriteString(fmt.Sprintf("#EXT-X-TARGETDURATION:%s\n", w.header.TargetDuration))
	for _, item := range w.items {
		file.WriteString(fmt.Sprintf("#EXTINF:%s\n", item.Extinf))
		file.WriteString(fmt.Sprintf("%s\n", item.Name))
	}
	file.WriteString(w.end)
	log.Printf("writer m3u8 file success\n")
}

func (w *m3u8Writer) WriteHeader(tag, value *string) {
	if strings.HasPrefix(*tag, "#EXT-X-VERSION") {
		w.header.Version = *value
	} else if strings.HasPrefix(*tag, "#EXT-X-MEDIA-SEQUENCE") {
		w.header.Sequence = *value
	} else if strings.HasPrefix(*tag, "#EXT-X-TARGETDURATION") {
		w.header.TargetDuration = *value
	} else {
		log.Printf("Unknown header tag: %s\n", *tag)
	}
}

func (w *m3u8Writer) WriteItem(item *item) {
	w.items = append(w.items, item)
}

func (w *m3u8Writer) GetLastItemName() (string, error) {
	if len(w.items) == 0 {
		return "", fmt.Errorf("no item exist")
	}
	return w.items[len(w.items)-1].Name, nil
}
