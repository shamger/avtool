package flv

import (
	"bytes"
	"encoding/binary"
	"strconv"
)

const (
	TagHeaderSize = 15
	FlvHeaderSize = 9
)

const (
	TagType_None  = 0x00
	TagType_Audio = 0x08
	TagType_Video = 0x09
	TagType_Meta  = 0x12
)

type TagHeader struct {
	PrevSize  uint32 //[4]byte
	TagType   byte
	DataSize  uint32 //[3]byte
	TimeStamp uint32 //[3]byte
	TimeExtra byte
	StreamId  [3]byte
}

func (t *TagHeader) GetBytes() []byte {
	tag := make([]byte, TagHeaderSize)
	prevSizeb := make([]byte, 4)
	binary.BigEndian.PutUint32(prevSizeb, t.PrevSize)
	copy(tag[0:4], prevSizeb)
	tag[4] = t.TagType

	dataSizeb := make([]byte, 4)
	binary.BigEndian.PutUint32(dataSizeb, t.DataSize)
	copy(tag[5:8], dataSizeb[:3])

	timeStampb := make([]byte, 4)
	binary.BigEndian.PutUint32(timeStampb, t.TimeStamp)
	copy(tag[8:11], timeStampb[:3])

	tag[11] = t.TimeExtra
	copy(tag[12:15], t.StreamId[:])
	return tag
}

func (t *TagHeader) GetStr() string {
	str := "PrevSize: " + strconv.FormatUint(uint64(t.PrevSize), 10) + "\n"
	str += "TagType: " + strconv.FormatUint(uint64(t.TagType), 10) + "\n"
	str += "DataSize: " + strconv.FormatUint(uint64(t.DataSize), 10) + "\n"
	str += "TimeStamp: " + strconv.FormatUint(uint64(t.TimeStamp), 10) + "\n"
	str += "============"
	return str
}

type FlvHeader struct { // 包括flv头和ScriptTag
	Header          [FlvHeaderSize]byte
	ScriptTagHeader TagHeader
	ScriptTagData   []byte
	Meta            map[string]interface{}
	DebugOrder      []string
}

func (f *FlvHeader) GetBytes() []byte {
	var buf bytes.Buffer
	//f.Header = [FlvHeaderSize]byte{'F', 'L', 'V', 0x01, 0x01, 0x00, 0x00, 0x00, 0x09}
	buf.Write(f.Header[:])
	if _, ok := f.Meta["duration"]; !ok {
		f.Meta["duration"] = 0.0
	}
	if _, ok := f.Meta["lasttimestamp"]; !ok {
		f.Meta["lasttimestamp"] = 0.0
	}

	metab := NewAmfEncoderDecoder().EncodeMetaData(f.Meta, f.DebugOrder)
	f.ScriptTagHeader.DataSize = uint32(len(metab))

	scriptTagHeaderb := f.ScriptTagHeader.GetBytes()
	buf.Write(scriptTagHeaderb)

	f.ScriptTagData = metab
	buf.Write(f.ScriptTagData)

	return buf.Bytes()
}
