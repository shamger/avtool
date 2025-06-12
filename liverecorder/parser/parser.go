package parser

import "liverecorder/parser/douyin"

type Parser interface {
	GetStreamUrl() (string, error)
}

func NewParser(url string, format string) Parser {
	d := &douyin.LiveRoom{
		LiveUrl:    url,
		FileFormat: format,
	}
	return d
}
