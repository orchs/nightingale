package ctf

import (
	"github.com/BurntSushi/toml"
)

var httpResponseInfo = CatConfInfo{
	Name:    "http_response",
	Indexes: []string{"instances.targets"},
	Toml: string(`
# # collect interval
interval = 15

[[instances]]
targets = [
     "https://teams.microsoft.com/",
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

type httpResponseToml struct {
	Interval  int                    `toml:"interval"`
	Instances []httpResponseInstance `toml:"instances"`
}

type httpResponseInstance struct {
	Targets []string `toml:"targets"`
}

func (t httpResponseToml) GetConfIndex(tomlStr string) ([]cIRecord, error) {
	var c httpResponseToml
	var cir []cIRecord

	_, err := toml.Decode(tomlStr, &c)
	if err != nil {
		return cir, err
	}

	for _, v := range httpResponseInfo.Indexes {
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
