package hls

// #EXTM3U
// #EXT-X-VERSION:3
// #EXT-X-MEDIA-SEQUENCE:9111
// #EXT-X-TARGETDURATION:5
// #EXTINF:5.000,
// pull-spe-l3.douyincdn.com_stream-7514824030164126514_thirdor4-1749727386951.ts
// #EXTINF:5.000,
// pull-spe-l3.douyincdn.com_stream-7514824030164126514_thirdor4-1749727391970.ts
// #EXTINF:5.000,
// pull-spe-l3.douyincdn.com_stream-7514824030164126514_thirdor4-1749727396987.ts
// #EXT-X-ENDLIST

type header struct {
	Version        string
	Sequence       string
	TargetDuration string
}

type item struct {
	Extinf string
	Name   string
}
