package flv_helper

import (
	"bytes"
	"testing"
)

func TestGetTimestamp(t *testing.T) {
	flvTag := &FlvTag{
		Header: *bytes.NewBuffer([]byte{0x09, 0x00, 0x13, 0x6c, 0x00, 0x4e, 0x0f, 0x01, 0x00, 0x00, 0x00}),
		Data:   *bytes.NewBuffer([]byte{}),
	}

	if flvTag.GetTimestamp() != 19983 {
		t.Error("GetTimestamp() failed, 19983 expected, but got ", flvTag.GetTimestamp())
	}
}
