module liverecorder

go 1.24.1

replace flvdumper => ../flvdumper
replace hlsdumper => ../hlsdumper
replace flvrewriter => ../flvrewriter

require github.com/tidwall/gjson v1.18.0
require flvdumper v0.0.0
require flvrewriter v0.0.0
require hlsdumper v0.0.0

require (
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.1 // indirect
)
