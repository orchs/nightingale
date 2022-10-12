package ctf

import (
	"github.com/BurntSushi/toml"
)

var redisInfo = CatConfInfo{
	Name:    "redis",
	Indexes: []string{"instances.address"},
	Toml: string(`
interval = 15
[[instances]]
address = "127.0.0.1:6379"
username = ""
password = ""
pool_size = 2
commands = [
    {command = ["get", "sample-key1"], metric = "custom_metric_name1"},
    {command = ["get", "sample-key2"], metric = "custom_metric_name2"}
]
interval_times = 1
# important! use global unique string to specify instance
# labels = { instance="n9e-10.2.3.4:6379" }
`,
	),
}

type redisToml struct {
	Interval  int             `toml:"interval"`
	Instances []redisInstance `toml:"instances"`
}

type redisInstance struct {
	Address string `toml:"address"`
}

func (t redisToml) GetConfIndex(tomlStr string) ([]cIRecord, error) {
	var c redisToml
	var cir []cIRecord

	_, err := toml.Decode(tomlStr, &c)
	if err != nil {
		return cir, err
	}

	for _, v := range redisInfo.Indexes {
		switch v {
		case "instances.address":
			for _, i := range c.Instances {
				cir = append(cir, cIRecord{v, i.Address})
			}
		}
	}
	return cir, nil
}
