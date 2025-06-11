package main

import (
	"flvdumper/dumper"
	"flvrewriter/flv"
	"fmt"
	"liverecorder/parser"
	"os"
	"strings"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Printf("Usage: %s <live_url> <output_filename.flv>\n", os.Args[0])
		return
	}
	liveUrl := os.Args[1]
	outputFilename := os.Args[2]
	if !strings.HasSuffix(outputFilename, ".flv") {
		fmt.Printf("output_filename must be .flv\n")
		return
	}

	streamUrl, err := parser.NewParser(liveUrl).GetStreamUrl()
	if err != nil {
		fmt.Printf("GetStreamUrl failed: %v\n", err)
		return
	}

	fmt.Printf("Got StreamUrl: %s\n", streamUrl)

	dumper.Process(streamUrl, outputFilename, flv.WriteType_Default)
	fmt.Printf("Stop recording.\n")
}
