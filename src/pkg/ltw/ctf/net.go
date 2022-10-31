package ctf

import (
	"github.com/BurntSushi/toml"
)

var netInfo = CatConfInfo{
	Name:    "net",
	Indexes: []string{"instances.targets"},
	Toml: string(`
# # collect interval
# interval = 15

# # whether collect protocol stats on Linux
# collect_protocol_stats = false

# # setting interfaces will tell categraf to gather these explicit interfaces
# interfaces = ["eth0"]
`,
	),
}

type netToml struct {
	Interval int `toml:"interval"`
}

func (t netToml) GetConfIndex(tomlStr string) ([]cIRecord, error) {
	var c netToml
	var cir []cIRecord

	_, err := toml.Decode(tomlStr, &c)
	if err != nil {
		return cir, err
	}

	for _, v := range netInfo.Indexes {
		switch v {
		case "interval":
			cir = append(cir, cIRecord{v, string(c.Interval)})
		}
	}
	return cir, nil
}
