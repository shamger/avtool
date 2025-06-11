package dumper

import (
	"context"
	"flvrewriter/flv"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func Process(inputUrl, outputFile string, writeType int) {
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

func dumpFile(flvWriter *flv.FlvWriter, inputBuf []byte) {
	flvWriter.Write(inputBuf)
}
