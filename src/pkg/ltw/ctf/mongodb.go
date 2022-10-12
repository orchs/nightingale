package ctf

import (
	"github.com/BurntSushi/toml"
)

var mongodbInfo = CatConfInfo{
	Name:    "mongodb",
	Indexes: []string{"instances.mongodb_uri"},
	Toml: string(`
[[instances]]
# log level, enum: panic, fatal, error, warn, warning, info, debug, trace, defaults to info.
log_level = "info"
labels = { instance="mongo-ope-cpe_prod" }
mongodb_uri = "mongodb://127.0.0.1:27017"
username = "username@Bj"
password = "password@Bj"
direct_connect = true
collect_all = true
interval_times = 1
compatible_mode = true
[[instances]]
# log level, enum: panic, fatal, error, warn, warning, info, debug, trace, defaults to info.
log_level = "info"
labels = { instance="mongo-ope-cpe_prod" }
mongodb_uri = "mongodb://127.0.0.1:27017"
username = "username@Bj"
password = "password@Bj"
direct_connect = true
collect_all = true
interval_times = 1
compatible_mode = true
`,
	),
}

type mongodbToml struct {
	Interval  int               `toml:"interval"`
	Instances []mongodbInstance `toml:"instances"`
}

type mongodbInstance struct {
	mongodbUri string `toml:"mongodb_uri"`
}

func (t mongodbToml) GetConfIndex(tomlStr string) ([]cIRecord, error) {
	var c mongodbToml
	var cir []cIRecord

	_, err := toml.Decode(tomlStr, &c)
	if err != nil {
		return cir, err
	}

	for _, v := range mongodbInfo.Indexes {
		switch v {
		case "instances.mongodb_uri":
			for _, i := range c.Instances {
				cir = append(cir, cIRecord{v, i.mongodbUri})
			}
		}
	}
	return cir, nil
}
