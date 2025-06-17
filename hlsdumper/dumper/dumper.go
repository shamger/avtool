package dumper

import (
	"context"
	"hlsdumper/hls"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

const (
	suffix_M3U8 = ".m3u8"
)

func Process(inputUrl, outputFile string) {
	var wg sync.WaitGroup
	handleDumpHls(&wg, inputUrl, outputFile)
	wg.Wait()
	log.Printf("all processor have been done\n")
}

func handleDumpHls(wg *sync.WaitGroup, inputUrl, outputFile string) {
	if !strings.HasSuffix(outputFile, suffix_M3U8) {
		log.Fatalf(`Output file must have .m3u8 extension, got: %s`, outputFile)
	}
	directory := strings.TrimSuffix(outputFile, suffix_M3U8)

	if _, err := os.Stat(directory); !os.IsNotExist(err) {
		if err := os.RemoveAll(directory); err != nil {
			log.Fatalf(`Failed to remove output directory: %v`, err)
		}
	}
	if err := os.Mkdir(directory, 0755); err != nil {
		log.Fatalf(`Failed to create output directory: %v`, err)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	parser := hls.NewM3u8Parser(ctx, wg, directory, outputFile)
	ticker := time.NewTicker(time.Duration(2 * time.Second))
readLoop:
	for {
		select {
		case <-sigCh:
			log.Printf("Received signal, exiting...")
			break readLoop
		default:
			client := &http.Client{Timeout: 3 * time.Second}
			resp, err := client.Get(inputUrl)
			if err != nil {
				log.Fatalf("get m3u8 content failed:%v", err)
				return
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				log.Fatalf("http error:%s", resp.Status)
				return
			}

			contentb, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Fatalf("get http response body failed:%v", err)
				return
			}
			log.Printf("get m3u8 content success\n%s\n%s\n", strconv.Quote(string(contentb)), contentb)
			if newurl, err := parser.Parse(string(contentb), inputUrl); err != nil {
				log.Fatalf("parse m3u8 failed:%v", err)
				return
			} else if len(newurl) != 0 {
				log.Printf("need updata m3u8 url: %s=>%s\n", inputUrl, newurl)
				inputUrl = newurl
				continue
			} else {
				// waitting for next ticker
				<-ticker.C
			}
		}
	}
}
