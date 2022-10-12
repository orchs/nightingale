package ctf

import (
	"github.com/BurntSushi/toml"
)

var procstatInfo = CatConfInfo{
	Name:    "procstat",
	Indexes: []string{"instances.search_exec_substring"},
	Toml: string(`
# # collect interval
interval = 15

 [[instances]]
# # executable name (ie, pgrep <search_exec_substring>)
search_exec_substring = "popagent"
interval_times = 2
mode = "irix"
gather_total = true
gather_per_pid = false
gather_more_metrics = [
     "threads",
     "fd",
     "io",
     "uptime",
     "cpu",
     "mem",
     "limit"
 ]
 
[[instances]]
search_exec_substring = "sshd"
interval_times = 2
`,
	),
}

type procstatToml struct {
	Interval  int                `toml:"interval"`
	Instances []procstatInstance `toml:"instances"`
}

type procstatInstance struct {
	SearchExecSubstring string `toml:"search_exec_substring"`
}

func (t procstatToml) GetConfIndex(tomlStr string) ([]cIRecord, error) {
	var c procstatToml
	var cir []cIRecord

	_, err := toml.Decode(tomlStr, &c)
	if err != nil {
		return cir, err
	}

	for _, v := range procstatInfo.Indexes {
		switch v {
		case "instances.search_exec_substring":
			for _, i := range c.Instances {
				cir = append(cir, cIRecord{v, i.SearchExecSubstring})
			}
		}
	}
	return cir, nil
}
