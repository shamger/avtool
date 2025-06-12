package main

import (
	"flvdumper/dumper"
	"flvrewriter/flv"
	"fmt"
	"liverecorder/parser"
	"os"
	"strings"
)

const (
	FileFormat_FLV = "flv"
	FileFormat_HLS = "hls"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Printf("Usage: %s <live_url> <output_filename.flv> [<format>]\n", os.Args[0])
		return
	}
	liveUrl := os.Args[1]
	outputFilename := os.Args[2]
	format := FileFormat_FLV
	if len(os.Args) == 4 {
		switch os.Args[3] {
		case FileFormat_HLS:
			format = FileFormat_HLS
		case FileFormat_FLV:
			fallthrough
		default:
			format = FileFormat_FLV
		}
	}
	if format == FileFormat_FLV {
		if !strings.HasSuffix(outputFilename, ".flv") {
			fmt.Printf("output_filename must be .flv\n")
			return
		}
	}
	if format == FileFormat_HLS {
		if !strings.HasSuffix(outputFilename, ".m3u8") {
			fmt.Printf("output_filename must be .m3u8\n")
			return
		}
	}

	streamUrl, err := parser.NewParser(liveUrl, format).GetStreamUrl()
	if err != nil {
		fmt.Printf("GetStreamUrl failed: %v\n", err)
		return
	}

	fmt.Printf("Got StreamUrl: %s\n", streamUrl)

	dumper.Process(streamUrl, outputFilename, flv.WriteType_Queue)
	fmt.Printf("Stop recording.\n")
}
