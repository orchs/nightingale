package ctf

import (
	"github.com/BurntSushi/toml"
)

var pingInfo = CatConfInfo{
	Name:    "ping",
	Indexes: []string{"instances.targets"},
	Toml: string(`
# # collect interval
interval = 15

[[instances]]
# send ping packets to
targets = [
     "172.19.16.50",
]

# # append some labels for series
labels = { tag="test", threshold="15"}

# # interval = global.interval * interval_times
interval_times = 2

## Number of ping packets to send per interval.  Corresponds to the "-c"
## option of the ping command.
count = 100

## Time to wait between sending ping packets in seconds.  Operates like the
## "-i" option of the ping command.
ping_interval = 0.1

## If set, the time to wait for a ping response in seconds.  Operates like
## the "-W" option of the ping command.
timeout = 1

## Interface or source targets to send ping from.  Operates like the -I or -S
## option of the ping command.
# interface = ""

## Use only IPv6 targetses when resolving a hostname.
# ipv6 = false

## Number of data bytes to be sent. Corresponds to the "-s"
## option of the ping command.
# size = 56
`,
	),
}

type pingToml struct {
	Interval  int            `toml:"interval"`
	Instances []pingInstance `toml:"instances"`
}

type pingInstance struct {
	Targets []string `toml:"targets"`
}

func (t pingToml) GetConfIndex(tomlStr string) ([]cIRecord, error) {
	var c pingToml
	var cir []cIRecord

	_, err := toml.Decode(tomlStr, &c)
	if err != nil {
		return cir, err
	}

	for _, v := range pingInfo.Indexes {
		switch v {
		case "instances.targets":
			for _, i := range c.Instances {
				for _, t := range i.Targets {
					cir = append(cir, cIRecord{v, t})
				}
			}
		}
	}
	return cir, err
}
