package main

import (
	"fmt"
	"liverecorder/parser"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("Usage: %s <live_url>\n", os.Args[0])
		return
	}
	liveUrl := os.Args[1]

	streamUrl, err := parser.NewParser(liveUrl).GetStreamUrl()
	if err != nil {
		fmt.Printf("GetStreamUrl failed: %v\n", err)
		return
	}

	fmt.Printf("Got StreamUrl: %s\n", streamUrl)
}
