package flv

import (
	"bytes"
	"encoding/binary"
	"log"
	"math"
	"reflect"
	"strings"
	"unsafe"
)

const (
	Amf0Type_Number      = 0x00
	Amf0Type_Boolean     = 0x01
	Amf0Type_String      = 0x02
	Amf0Type_Object      = 0x03
	Amf0Type_Null        = 0x05
	Amf0Type_ECMAArray   = 0x08
	Amf0Type_ObjectEnd   = 0x09
	Amf0Type_StrictArray = 0x0A
	Amf0Type_Date        = 0x0B
	Amf0Type_LongString  = 0x0C
	Amf0Type_XMLDoc      = 0x0F
	// ...
)

type AmfEncoderDecoder struct {
	readHead   int
	objs       map[string]interface{}
	debugOrder []string
}

func NewAmfEncoderDecoder() *AmfEncoderDecoder {
	return &AmfEncoderDecoder{
		readHead:   0,
		objs:       make(map[string]interface{}),
		debugOrder: make([]string, 0),
	}
}

func reverse(in []byte) []byte {
	out := make([]byte, len(in))
	for i, j := 0, len(in)-1; i <= j; i, j = i+1, j-1 {
		out[i], out[j] = in[j], in[i]
	}
	return out
}

func (amf *AmfEncoderDecoder) DecodeKey(buff []byte) string {
	flip := make([]byte, 2) // two bytes denoting their length
	flip[0] = buff[amf.readHead]
	amf.readHead++
	flip[1] = buff[amf.readHead]
	amf.readHead++
	klen := binary.BigEndian.Uint16(flip)
	name := string(buff[amf.readHead : amf.readHead+int(klen)])
	amf.readHead += int(klen)
	return name
}

func (amf *AmfEncoderDecoder) DecodeVal(buff []byte) interface{} {
	amfType := buff[amf.readHead]
	amf.readHead++

	switch amfType {
	case Amf0Type_String:
		return amf.DecodeKey(buff)
	case Amf0Type_Number:
		var f float64
		flip := make([]byte, unsafe.Sizeof(f))
		copy(flip, buff[amf.readHead:amf.readHead+len(flip)])
		num := math.Float64frombits(binary.BigEndian.Uint64(flip))
		amf.readHead += len(flip)
		return num
	case Amf0Type_Boolean:
		b := buff[amf.readHead]
		amf.readHead++
		return b != 0
	case Amf0Type_ObjectEnd:
		return nil
	default:
		log.Fatalf("Unknown amf0 type: %v", amfType)
		return nil
	}
}

func (amf *AmfEncoderDecoder) EncodeKey(key string) []byte {
	ret := make([]byte, 2+len(key))
	strSize := uint16(len(key))
	binary.BigEndian.PutUint16(ret, strSize)
	copy(ret[2:], key)
	return ret
}

func (amf *AmfEncoderDecoder) EncodeVal(val interface{}) []byte {
	if reflect.TypeOf(val) == reflect.TypeOf(float64(0)) {
		num := val.(float64)
		ret := make([]byte, 1+unsafe.Sizeof(num))
		ret[0] = Amf0Type_Number
		binary.BigEndian.PutUint64(ret[1:], math.Float64bits(num))
		return ret
	} else if reflect.TypeOf(val) == reflect.TypeOf(string("")) {
		str := val.(string)
		ret := make([]byte, 3+len(str))
		ret[0] = Amf0Type_String
		strSize := uint16(len(str))
		binary.BigEndian.PutUint16(ret[1:], strSize)
		copy(ret[3:], str)
		return ret
	} else if reflect.TypeOf(val) == reflect.TypeOf(bool(false)) {
		byteVal := byte(0)
		if val.(bool) {
			byteVal = 1
		}
		ret := make([]byte, 2)
		ret[0] = Amf0Type_Boolean
		ret[1] = byteVal
		return ret
	} else {
		log.Fatalf("Unknown type: %v", reflect.TypeOf(val))
		return nil
	}
}

func (amf *AmfEncoderDecoder) EncodeMetaData(meta map[string]interface{}, order []string) []byte {
	var buf bytes.Buffer
	onMetaData := "onMetaData"
	buf.WriteByte(Amf0Type_String)
	strSize := uint16(len(onMetaData))
	strSizeb := make([]byte, unsafe.Sizeof(strSize))
	binary.BigEndian.PutUint16(strSizeb, strSize)
	buf.Write(strSizeb)
	buf.WriteString(onMetaData)

	buf.WriteByte(Amf0Type_ECMAArray)
	arrSize := uint16(len(meta))
	arrSizeb := make([]byte, unsafe.Sizeof(arrSize))
	binary.BigEndian.PutUint16(arrSizeb, arrSize)
	buf.Write(arrSizeb)

	if len(order) > 0 {
		for _, key := range order {
			val := meta[key]
			if len(strings.TrimSpace(key)) > 0 && val != nil {
				keyBytes := amf.EncodeKey(key)
				buf.Write(keyBytes)
				valBytes := amf.EncodeVal(val)
				buf.Write(valBytes)
			}
		}
	} else {
		for key, val := range meta {
			if len(strings.TrimSpace(key)) > 0 && val != nil {
				keyBytes := amf.EncodeKey(key)
				buf.Write(keyBytes)
				valBytes := amf.EncodeVal(val)
				buf.Write(valBytes)
			}
		}
	}
	buf.WriteByte(0x00)
	buf.WriteByte(0x00)
	buf.WriteByte(Amf0Type_ObjectEnd)
	return buf.Bytes()
}

func (amf *AmfEncoderDecoder) DecodeMetaData(buff []byte) map[string]interface{} {
	keyval := make(map[string]interface{})
	amf.readHead = 0
	onMetaData, ok := amf.DecodeVal(buff).(string)
	if !ok {
		log.Fatalf("onMetaData is not string")
	} else {
		log.Printf("onMetaData: %s", onMetaData)
	}

	amftype := buff[amf.readHead]
	amf.readHead++
	if amftype == Amf0Type_ECMAArray {
		alen := make([]byte, unsafe.Sizeof(int(0)))
		copy(alen, buff[amf.readHead:amf.readHead+len(alen)])
		amf.readHead += len(alen)
		arrayLen := binary.BigEndian.Uint32(alen)
		log.Printf("onMetaData arrayLen: %d", arrayLen)
	} else if amftype == Amf0Type_Object {
		log.Fatalf("onMetaData is not ECMAArray but Object")
	} else {
		log.Fatalf("amf type of onMetaData is not ECMAArray or Object: %d", amftype)
	}

	for amf.readHead < len(buff) {
		key := amf.DecodeKey(buff)
		amf.debugOrder = append(amf.debugOrder, key)
		val := amf.DecodeVal(buff)
		log.Printf("decode meta data key: %s, val: %v", key, val)
		keyval[key] = val
	}
	return keyval
}
