package main

import (
	"flvrewriter/flv"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

func main() {
	if len(os.Args) < 3 {
		log.Fatalf("Usage: %s <http-flv-url> <output.flv> [<startTagIndex> <endTagIndex>]", os.Args[0])
		return
	}

	inputUrl := os.Args[1]
	outputFile := os.Args[2]

	// 打开输出文件
	flvWriter := flv.Open(outputFile)
	defer flvWriter.Close()
	if len(os.Args) == 5 {
		flvWriter.PrintTagStartIdx, _ = strconv.Atoi(os.Args[3])
		flvWriter.PrintTagEndIdx, _ = strconv.Atoi(os.Args[4])
	}

	// 准备读区url
	resp, err := http.Get(inputUrl)
	if err != nil {
		log.Fatalf("Failed to open url: %s", inputUrl)
	}
	defer resp.Body.Close()

	// 设置信号处理函数
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	// 读取flv数据
	readBuffer := make([]byte, 4096)
	for {
		select {
		case <-sigCh:
			log.Fatalf("User stopped")
			return
		default:
			if bytesRead, err := io.ReadFull(resp.Body, readBuffer); err == io.EOF {
				flvWriter.Write(readBuffer[:bytesRead])
				log.Fatalf("End of file")
				return
			} else if err != nil {
				log.Fatalf("Failed to read file: %s", err.Error())
				return
			} else if bytesRead != len(readBuffer) {
				flvWriter.Write(readBuffer[:bytesRead])
			} else {
				flvWriter.Write(readBuffer)
			}
		}
	}
}
