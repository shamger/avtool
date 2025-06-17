package hls

import "testing"

func TestNotifyDownloader(t *testing.T) {
	parser := &m3u8Parser{
		writer:         nil,
		itemDownloader: nil,
	}
	m3u8Url := "http://pull-tsl-q11.douyincdn.com/fantasy-bak/dev-stream1749630189/timeshift.m3u8?expire=1749718305&sign=27ff5756f5c9b43cd66f7bf84b20a125&starttime=20250611165800&endtime=20250611171000"
	err := parser.notifyDownloader("testname.ts", m3u8Url)
	if err != nil {
		t.Error(err)
	}
}
