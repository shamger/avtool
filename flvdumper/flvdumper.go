package main

import (
	"flvdumper/dumper"
	"flvrewriter/flv"
	"fmt"
	"log"
	"os"
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

	dumper.Process(inputUrl, outputFile, writeType)

	time.Sleep(time.Duration(time.Second))
	log.Printf("master exist")
}
