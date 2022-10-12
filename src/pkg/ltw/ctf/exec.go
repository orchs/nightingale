package ctf

import (
	"github.com/BurntSushi/toml"
)

var execInfo = CatConfInfo{
	Name:    "exec",
	Indexes: []string{"instances.commands"},
	Toml: string(`
# # collect interval
interval = 30

[[instances]]
# # commands, support glob
commands = [
     "/opt/categraf/scripts/*.sh"
]
interval_times = 2
timeout = 60
# # mesurement,labelkey1=labelval1,labelkey2=labelval2 field1=1.2,field2=2.3
data_format = "influx"
`,
	),
}

type execToml struct {
	Interval  int            `toml:"interval"`
	Instances []execInstance `toml:"instances"`
}
type execInstance struct {
	Commands []string `toml:"commands"`
}

func (t execToml) GetConfIndex(tomlStr string) ([]cIRecord, error) {
	var c execToml
	var cir []cIRecord

	_, err := toml.Decode(tomlStr, &c)
	if err != nil {
		return cir, err
	}

	for _, v := range execInfo.Indexes {
		switch v {
		case "instances.commands":
			for _, i := range c.Instances {
				for _, t := range i.Commands {
					cir = append(cir, cIRecord{v, t})
				}
			}
		}
	}
	return cir, nil
}
