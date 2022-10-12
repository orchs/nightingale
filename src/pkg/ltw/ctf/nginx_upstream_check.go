package ctf

import (
	"github.com/BurntSushi/toml"
)

var nginxUpstreamCheckInfo = CatConfInfo{
	Name:    "http_response",
	Indexes: []string{"instances.targets"},
	Toml: string(`
interval = 15

[[instances]]
targets = [
    "http://127.0.0.1/status?format=json",
]
labels = { index="nginx" }
interval_times = 1
interface = "eth0"
method = "GET"
timeout = "5s"
follow_redirects = true
`,
	),
}

type nginxUpstreamCheckToml struct {
	Interval  int                          `toml:"interval"`
	Instances []nginxUpstreamCheckInstance `toml:"instances"`
}

type nginxUpstreamCheckInstance struct {
	Targets []string `toml:"targets"`
}

func (t nginxUpstreamCheckToml) GetConfIndex(tomlStr string) ([]cIRecord, error) {
	var c nginxUpstreamCheckToml
	var cir []cIRecord

	_, err := toml.Decode(tomlStr, &c)
	if err != nil {
		return cir, err
	}

	for _, v := range nginxUpstreamCheckInfo.Indexes {
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
