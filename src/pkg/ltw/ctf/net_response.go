package ctf

import (
	"github.com/BurntSushi/toml"
)

var netResponseInfo = CatConfInfo{
	Name:    "net_response",
	Indexes: []string{"instances.targets"},
	Toml: string(`
# # collect interval
interval = 15

[[instances]]
targets = [
     "www.baidu.com:80",
]

# # append some labels for series
labels = { tag="XXX项目" }

# # interval = global.interval * interval_times
interval_times = 2

## Protocol, must be "tcp" or "udp"
## NOTE: because the "udp" protocol does not respond to requests, it requires
## a send/expect string pair (see below).
protocol = "tcp"

## Set timeout
timeout = "5s"

## Set read timeout (only used if expecting a response)
# read_timeout = "1s"

## The following options are required for UDP checks. For TCP, they are
## optional. The plugin will send the given string to the server and then
## expect to receive the given 'expect' string back.
## string sent to the server
# send = "ssh"
## expected string in answer
# expect = "ssh"
`,
	),
}

type netResponseToml struct {
	Interval  int                   `toml:"interval"`
	Instances []netResponseInstance `toml:"instances"`
}

type netResponseInstance struct {
	Targets []string `toml:"targets"`
}

func (t netResponseToml) GetConfIndex(tomlStr string) ([]cIRecord, error) {
	var c netResponseToml
	var cir []cIRecord

	_, err := toml.Decode(tomlStr, &c)
	if err != nil {
		return cir, err
	}

	for _, v := range netResponseInfo.Indexes {
		switch v {
		case "instances.targets":
			for _, i := range c.Instances {
				for _, t := range i.Targets {
					cir = append(cir, cIRecord{v, t})
				}
			}
		}
	}
	return cir, nil
}
