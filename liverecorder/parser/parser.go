package parser

type Parser interface {
	GetStreamUrl() (string, error)
}

func NewParser(url string) Parser {
	d := &douyin{
		liveUrl: url,
	}
	return d
}
