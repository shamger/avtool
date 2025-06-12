package flv_helper

import (
	"bytes"
	"log"
)

// 通过缓冲队列写入文件，需要解析flv tag data的帧类型来写入完整gop
type FlvTag struct {
	Header bytes.Buffer
	Data   bytes.Buffer // first 4bit: video-frame type; audio-audio format
}

func (f *FlvTag) IsKeyTag() bool {
	if f.Header.Bytes()[4] == 0x09 {
		b, err := f.Data.ReadByte()
		if err != nil {
			log.Fatalf("tag data error")
			return false
		}
		f.Data.UnreadByte()
		videoFrameType := (b >> 4) & 0x0F
		if videoFrameType == 1 {
			return true
		} else {
			return false
		}
	} else {
		return false
	}
}

func (f *FlvTag) IsVideoTag() bool {
	if f.Header.Bytes()[4] == 0x09 {
		return true
	} else {
		return false
	}
}

func (f *FlvTag) GetTimestamp() uint32 {
	timestamp := uint32(f.Header.Bytes()[5]&0x01)<<16 | uint32(f.Header.Bytes()[6]&0xFF)<<8 | uint32(f.Header.Bytes()[7]&0xFF)
	return timestamp
}
