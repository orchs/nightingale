package ctf

import (
	"github.com/BurntSushi/toml"
)

var dnsQueryInfo = CatConfInfo{
	Name:    "dns_query",
	Indexes: []string{"instances.servers", "instances.domains"},
	Toml: string(`
# # collect interval
interval = 15

[[instances]]
## Overseas: labels = { region="Overseas" }
labels = { region="ChinaMainland" }
interval_times = 2
## servers to query
## Overseas: servers = ["8.8.8.8", "1.1.1.1"]
servers = ["114.114.114.114", "223.5.5.5"]
domains = []
timeout = 5
`,
	),
}

type dnsQueryToml struct {
	Interval  int                `toml:"interval"`
	Instances []dnsQueryInstance `toml:"instances"`
}

type dnsQueryInstance struct {
	Servers []string `toml:"servers"`
	Domains []string `toml:"domains"`
}

func (t dnsQueryToml) GetConfIndex(tomlStr string) ([]cIRecord, error) {
	var c dnsQueryToml
	var cir []cIRecord

	_, err := toml.Decode(tomlStr, &c)
	if err != nil {
		return cir, err
	}

	for _, v := range dnsQueryInfo.Indexes {
		switch v {
		case "instances.servers":
			for _, i := range c.Instances {
				for _, t := range i.Servers {
					cir = append(cir, cIRecord{v, t})
				}
			}
		case "instances.domains":
			for _, i := range c.Instances {
				for _, t := range i.Domains {
					cir = append(cir, cIRecord{v, t})
				}
			}
		}
	}
	return cir, nil
}
