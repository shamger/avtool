package main

import (
	"context"
	"flvrewriter/flv"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Printf("Usage: %s <http-flv-url> <output.flv> [<option>]\n", os.Args[0])
		fmt.Printf("option:\n")
		fmt.Printf("	-directly	write file directly by flv rewriter\n")
		fmt.Printf("	-queue	write file by flv rewriter via queue\n")
		fmt.Printf("	-raw	write file in binary\n")
		return
	}
	option := "-directly"
	if len(os.Args) == 4 {
		option = os.Args[3]
	}
	writeType := flv.WriteType_Default
	switch option {
	case "-queue":
		writeType = flv.WriteType_Queue
	case "-raw":
		writeType = flv.WriteType_RawBin
	case "-directly":
		fallthrough
	default:
		writeType = flv.WriteType_Directly
	}

	inputUrl := os.Args[1]
	outputFile := os.Args[2]

	process(inputUrl, outputFile, writeType)

	time.Sleep(time.Duration(time.Second))
	log.Printf("master exist")
}

func dumpFile(flvWriter *flv.FlvWriter, inputBuf []byte) {
	flvWriter.Write(inputBuf)
}

func process(inputUrl, outputFile string, writeType int) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 打开输出文件
	flvWriter := flv.Open(ctx, outputFile, writeType)
	defer flvWriter.Close()

	// 准备读取url
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
readLoop:
	for {
		select {
		case <-sigCh:
			log.Printf("User stopped")
			break readLoop
		default:
			if bytesRead, err := io.ReadFull(resp.Body, readBuffer); err == io.EOF {
				dumpFile(flvWriter, readBuffer[:bytesRead])
				log.Printf("End of file")
				return
			} else if err != nil {
				log.Fatalf("Failed to read file: %s", err.Error())
				return
			} else if bytesRead != len(readBuffer) {
				dumpFile(flvWriter, readBuffer[:bytesRead])
			} else {
				dumpFile(flvWriter, readBuffer)
			}
		}
	}
	// directlyWriteFile
	// 到这里说明流没有结束就退出循环，如果最后一个tag还未写完整，需要回退
	//flvWriter.EraseLastBrokenTag()
}
