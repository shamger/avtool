package flv

import (
	"bytes"
	"errors"
	"io"
	"log"
	"os"
)

const BUF_LEN = 2048 // ?

type FlvWriter struct {
	fileName string
	outFile  *os.File

	readPos   int
	hasHeader bool
	hasOffset bool
	inputBuf  bytes.Buffer // _ms

	numVideo      int
	numAudio      int
	maxTimeStamp  int
	baseTimeStamp int

	header *FlvHeader
	curTag *TagHeader
}

func Open(filename string) *FlvWriter {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Fatalf("Failed to open file: %s", err.Error())
		return nil
	}
	return &FlvWriter{
		fileName: filename,
		outFile:  file,

		readPos:   0,
		hasHeader: false,
		hasOffset: false,
		inputBuf:  bytes.Buffer{},

		numVideo:      0,
		numAudio:      0,
		maxTimeStamp:  0,
		baseTimeStamp: 0,

		header: &FlvHeader{},
		curTag: &TagHeader{},
	}
}

func (w *FlvWriter) Write(buffer []byte) error {
	if w.outFile == nil {
		return errors.New("file not open")
	}
	if !w.hasHeader {
		w.grabHeader(buffer)
	} else if w.curTag.DataSize > 0 {
		w.readTagData(buffer)
	} else {
		w.parseTagHeader(buffer)
	}
	return nil
}

func (w *FlvWriter) Close() {
	// 更新metadata
	w.header.Meta["duration"] = w.maxTimeStamp / 1000.0
	w.header.Meta["lasttimestamp"] = w.maxTimeStamp
	header := w.header.GetBytes()
	// 重新定位到文件开头写header
	w.outFile.Seek(0, io.SeekStart)
	n, err := w.outFile.Write(header)
	if err != nil || n != len(header) {
		log.Fatalf("Failed to write header: %s, written: %dbytes", err.Error(), n)
	}
	// 关闭文件
	w.outFile.Close()
}

func (w *FlvWriter) grabHeader(buffer []byte) {
	w.inputBuf.Write(buffer)
	if w.inputBuf.Len() >= BUF_LEN { // TODO BUF_LEN是否合理?
		w.inputBuf.Read(w.header.Header[:])
		w.inputBuf.Read(w.header.ScriptTagHeader.PrevSize[:])
		if tagType, err := w.inputBuf.ReadByte(); err != nil {
			log.Fatalf("Failed to read tag type: %s", err.Error())
		} else {
			w.header.ScriptTagHeader.TagType = tagType
		}
	}
}

func (w *FlvWriter) parseTagHeader(buffer []byte) {

}

func (w *FlvWriter) readTagData(buffer []byte) {
}
