package flv

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"log"
	"os"
	"strconv"
)

const BUF_LEN = 2048 // NOTE: this MUST be large enough to get header + meta data...

type FlvWriter struct {
	PrintTagStartIdx int
	PrintTagEndIdx   int
	Option           string

	needWrite bool // 标记一个tag是否需要写入输出文件

	outFileName string
	outFile     *os.File

	tagIndex        int
	lastTagDataSize uint32
	tagDataReadPos  int
	hasHeader       bool
	hasOffset       bool
	inputBuf        bytes.Buffer // _ms

	numVideo      int
	numAudio      int
	maxTimeStamp  uint32
	baseTimeStamp uint32

	header *FlvHeader
	curTag *TagHeader
}

func Open(outFilename string) *FlvWriter {
	file, err := os.OpenFile(outFilename, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Fatalf("Failed to open file: %s", err.Error())
		return nil
	}
	return &FlvWriter{
		PrintTagStartIdx: 0,
		PrintTagEndIdx:   0,
		Option:           "-show",

		needWrite: true,

		outFileName: outFilename,
		outFile:     file,

		tagIndex:       0,
		tagDataReadPos: 0,
		hasHeader:      false,
		hasOffset:      false,
		inputBuf:       bytes.Buffer{},

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
	// 写入最后一个PrevSize
	prevSizeb := make([]byte, 4)
	prevSize := TagHeaderSize + w.lastTagDataSize - 4
	binary.BigEndian.PutUint32(prevSizeb, uint32(prevSize))
	if n, err := w.outFile.Write(prevSizeb); err != nil || n != len(prevSizeb) {
		log.Fatalf("Failed to write prev size: %s, written: %dbytes", err.Error(), n)
	}
	// 更新metadata
	w.header.Meta["duration"] = float64(w.maxTimeStamp) / 1000.0
	w.header.Meta["lasttimestamp"] = float64(w.maxTimeStamp) / 1.0
	header := w.header.GetBytes(false)
	// 重新定位到文件开头写header
	w.outFile.Seek(0, io.SeekStart)
	if n, err := w.outFile.Write(header); err != nil || n != len(header) {
		log.Fatalf("Failed to write header: %s, written: %dbytes", err.Error(), n)
	}
	// 关闭文件
	w.outFile.Close()
}

func (w *FlvWriter) GetDebugInfo() string {
	str := "Duration: " + strconv.FormatFloat(float64(w.maxTimeStamp/1000.0), 'f', 3, 64) + "s\n"
	str += "Num Audio: " + strconv.FormatInt(int64(w.numAudio), 10) + "\n"
	str += "Num Video: " + strconv.FormatInt(int64(w.numVideo), 10) + "\n"
	str += "Max Tag Index: " + strconv.FormatUint(uint64(w.tagIndex), 10) + "\n"
	return str
}

func (w *FlvWriter) grabHeader(buffer []byte) {
	w.inputBuf.Write(buffer)
	if w.inputBuf.Len() >= BUF_LEN {
		// flv header
		w.inputBuf.Read(w.header.Header[:])
		// script tag header
		prevSizeb := make([]byte, 4)
		if _, err := w.inputBuf.Read(prevSizeb); err != nil {
			log.Fatalf("Failed to read PrevSize: %v", err)
		} else {
			w.header.ScriptTagHeader.PrevSize = binary.BigEndian.Uint32(prevSizeb)
		}
		if tagType, err := w.inputBuf.ReadByte(); err != nil {
			log.Fatalf("Failed to read tag type: %v", err)
		} else {
			w.header.ScriptTagHeader.TagType = tagType
		}
		dataSizeb := make([]byte, 4)
		if _, err := w.inputBuf.Read(dataSizeb[1:]); err != nil {
			log.Fatalf("Failed to read data size: %v", err)
		} else {
			w.header.ScriptTagHeader.DataSize = binary.BigEndian.Uint32(dataSizeb)
		}
		timeStampb := make([]byte, 4)
		if _, err := w.inputBuf.Read(timeStampb[1:]); err != nil {
			log.Fatalf("Failed to read time stamp: %v", err)
		} else {
			w.header.ScriptTagHeader.TimeStamp = binary.BigEndian.Uint32(timeStampb)
		}
		if timeExta, err := w.inputBuf.ReadByte(); err != nil {
			log.Fatalf("Failed to read time extra: %v", err)
		} else {
			w.header.ScriptTagHeader.TimeExtra = timeExta
		}
		if _, err := w.inputBuf.Read(w.header.ScriptTagHeader.StreamId[:]); err != nil {
			log.Fatalf("Failed to read stream id: %v", err)
		}
		log.Printf("script tag header: %s", w.header.ScriptTagHeader.GetStr())

		// script tag data
		w.header.ScriptTagData = make([]byte, w.header.ScriptTagHeader.DataSize)
		if _, err := w.inputBuf.Read(w.header.ScriptTagData); err != nil {
			log.Fatalf("Failed to read script tag data: %v", err)
		}
		// 解析meta data备用
		amf := NewAmfEncoderDecoder()
		w.header.Meta = amf.DecodeMetaData(w.header.ScriptTagData)
		// w.header.DebugOrder = amf.GetDebugOrder() // only for debug, no adding any metadata
		// 写入flv header和script tag
		w.outFile.Write(w.header.GetBytes(false))

		// 准备开始写音视频tag
		w.hasHeader = true
		w.tagDataReadPos = 0
		rest := make([]byte, w.inputBuf.Len())
		w.inputBuf.Read(rest)
		w.inputBuf.Reset() // 清空输入缓冲
		w.Write(rest)
	}
}

func (w *FlvWriter) parseTagHeader(buffer []byte) {
	w.inputBuf.Write(buffer)
	if w.inputBuf.Len() >= TagHeaderSize {
		prevSizeb := make([]byte, 4)
		if _, err := w.inputBuf.Read(prevSizeb); err != nil {
			log.Fatalf("Failed to read tag prev size: %v", err)
		} else {
			w.curTag.PrevSize = binary.BigEndian.Uint32(prevSizeb)
		}
		if tagType, err := w.inputBuf.ReadByte(); err != nil {
			log.Fatalf("Failed to read tag type: %v", err)
		} else {
			w.curTag.TagType = tagType
		}
		dataSizeb := make([]byte, 4)
		if _, err := w.inputBuf.Read(dataSizeb[1:]); err != nil {
			log.Fatalf("Failed to read data size: %v", err)
		} else {
			w.curTag.DataSize = binary.BigEndian.Uint32(dataSizeb)
		}
		timeStampb := make([]byte, 4)
		if _, err := w.inputBuf.Read(timeStampb[1:]); err != nil {
			log.Fatalf("Failed to read time stamp: %v", err)
		} else {
			w.curTag.TimeStamp = binary.BigEndian.Uint32(timeStampb)
		}
		if timeExtra, err := w.inputBuf.ReadByte(); err != nil {
			log.Fatalf("Failed to read time extra: %v", err)
		} else {
			w.curTag.TimeExtra = timeExtra
		}
		if _, err := w.inputBuf.Read(w.curTag.StreamId[:]); err != nil {
			log.Fatalf("Failed to read stream id: %v", err)
		}

		// 读完tagHeader所有字段后，先处理option
		w.handleOption()

		if w.needWrite {
			// reset timestamp
			w.curTag.TimeStamp -= w.baseTimeStamp
			w.maxTimeStamp = func() uint32 {
				if w.maxTimeStamp > w.curTag.TimeStamp {
					return w.maxTimeStamp
				} else {
					return w.curTag.TimeStamp
				}
			}()
			// 跳过AAC头或者AVC头获取基准时戳，哪个先来用哪个作为基准
			if w.curTag.TagType == TagType_Audio {
				w.numAudio++
				if !w.hasOffset && w.numAudio == 2 {
					w.hasOffset = true
					log.Printf("Reset to base timestamp(%d): %d", w.curTag.TagType, w.curTag.TimeStamp)
					w.baseTimeStamp = w.curTag.TimeStamp
					w.curTag.TimeStamp = 0
					w.maxTimeStamp = 0
				}
			} else if w.curTag.TagType == TagType_Video {
				w.numVideo++
				if !w.hasOffset && w.numVideo == 2 {
					w.hasOffset = true
					log.Printf("Reset to base timestamp(%d): %d", w.curTag.TagType, w.curTag.TimeStamp)
					w.baseTimeStamp = w.curTag.TimeStamp
					w.curTag.TimeStamp = 0
					w.maxTimeStamp = 0
				}
			} else {
				log.Fatalf("unknown tag type:%d", w.curTag.TagType)
			}

			// tag header 写入文件
			tagHeader := w.curTag.GetBytes()
			w.outFile.Write(tagHeader)
		}

		w.tagIndex++
		// 重写剩余buffer
		rest := make([]byte, w.inputBuf.Len())
		w.inputBuf.Read(rest)
		w.inputBuf.Reset()
		w.Write(rest)
	}
}

func (w *FlvWriter) readTagData(buffer []byte) {
	toRead := func() int {
		if len(buffer) < int(w.curTag.DataSize)-w.tagDataReadPos {
			return len(buffer)
		} else {
			return int(w.curTag.DataSize) - w.tagDataReadPos
		}
	}()
	if w.needWrite {
		w.outFile.Write(buffer[:toRead])
	}
	w.tagDataReadPos += toRead
	if w.tagDataReadPos == int(w.curTag.DataSize) {
		if w.needWrite { // 确实有写入的才需要记录最后一个tag大小
			w.lastTagDataSize = w.curTag.DataSize
		}
		// 写完一个tag data
		w.curTag.TagType = 0
		w.curTag.DataSize = 0
		w.tagDataReadPos = 0
		w.Write(buffer[toRead:])
	}
}

func (w *FlvWriter) handleOption() {
	switch w.Option {
	case "-cp":
		if w.tagIndex >= w.PrintTagStartIdx && w.tagIndex < w.PrintTagEndIdx {
			w.needWrite = true
			return
		} else {
			if w.curTag.TagType != TagType_Audio && w.curTag.TagType != TagType_Video {
				w.needWrite = true
			} else {
				w.needWrite = false
			}
			return
		}
	case "-show":
		fallthrough
	default:
		if w.tagIndex >= w.PrintTagStartIdx && w.tagIndex < w.PrintTagEndIdx {
			log.Printf("writing tag head[%d]: \n%s", w.tagIndex, w.curTag.GetStr())
		}
		w.needWrite = true
		return
	}

}
