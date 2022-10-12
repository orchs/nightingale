package ctf

import (
	"github.com/BurntSushi/toml"
)

var redisSentinelInfo = CatConfInfo{
	Name:    "redisSentinel",
	Indexes: []string{"instances.servers"},
	Toml: string(`
interval = 15
[[instances]]
# [protocol://][:password]@servers[:port]
# e.g. servers = ["tcp://localhost:26379"]
servers = []
interval_times = 1
labels = {}
`,
	),
}

type redisSentinelToml struct {
	Interval  int                     `toml:"interval"`
	Instances []redisSentinelInstance `toml:"instances"`
}

type redisSentinelInstance struct {
	Servers []string `toml:"servers"`
}

func (t redisSentinelToml) GetConfIndex(tomlStr string) ([]cIRecord, error) {
	var c redisSentinelToml
	var cir []cIRecord

	_, err := toml.Decode(tomlStr, &c)
	if err != nil {
		return cir, err
	}

	for _, v := range redisSentinelInfo.Indexes {
		switch v {
		case "instances.servers":
			for _, i := range c.Instances {
				for _, t := range i.Servers {
					cir = append(cir, cIRecord{v, t})
				}
			}
		}
	}

	return cir, nil
}
