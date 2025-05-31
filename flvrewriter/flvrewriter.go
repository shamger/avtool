package main

import (
	"flvrewriter/flv"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
)

func main() {
	if len(os.Args) == 2 && os.Args[1] == "-h" {
		fmt.Printf("Usage: %s <input.flv> <output.flv> [<option> <startTagIndex> <endTagIndex>]\n", os.Args[0])
		fmt.Printf("option:\n")
		fmt.Printf("	-show	Show tag indexes between <startTagIndex> and <endTagIndex>\n")
		fmt.Printf("	-cp	Copy tag indexes between <startTagIndex> and <endTagIndex>\n")
		return
	}
	if len(os.Args) < 3 {
		fmt.Printf("Usage: %s <input.flv> <output.flv> [<option> <startTagIndex> <endTagIndex>]\n", os.Args[0])
		return
	}
	inputFile := os.Args[1]
	outputFile := os.Args[2]

	flvWriter := flv.Open(outputFile)
	defer flvWriter.Close()

	if len(os.Args) == 6 {
		flvWriter.Option = os.Args[3]
		flvWriter.PrintTagStartIdx, _ = strconv.Atoi(os.Args[4])
		flvWriter.PrintTagEndIdx, _ = strconv.Atoi(os.Args[5])
	}

	inFile, err := os.OpenFile(inputFile, os.O_RDONLY, 0666)
	if err != nil {
		log.Fatalf("Failed to open file: %s", err.Error())
		return
	}
	defer inFile.Close()

	readBuffer := make([]byte, 4096)
	for {
		if bytesRead, err := inFile.Read(readBuffer); err == io.EOF {
			flvWriter.Write(readBuffer[:bytesRead])
			break
		} else if err != nil {
			log.Fatalf("Failed to read file: %s", err.Error())
			return
		} else if bytesRead != len(readBuffer) {
			flvWriter.Write(readBuffer[:bytesRead])
		} else {
			flvWriter.Write(readBuffer)
		}
	}
	log.Printf("Finish writing file: %s\n%s", outputFile, flvWriter.GetDebugInfo())
}
