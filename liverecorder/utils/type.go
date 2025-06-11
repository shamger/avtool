package utils

import "net/url"

const (
	UserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/59.0.3071.115 Safari/537.36"
)

type StreamUrlInfo struct {
	Url                *url.URL
	Name               string
	Description        string
	Resolution         int
	Vbitrate           int
	HeadersForDownload map[string]string
}
