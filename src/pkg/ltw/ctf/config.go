package ctf

import (
	"github.com/BurntSushi/toml"
)

var configInfo = CatConfInfo{
	Name:    "config",
	Indexes: []string{"writers.url"},
	Toml: string(`
[global]
# whether print configs
print_configs = false

# add label(agent_hostname) to series
hostname = "$hostname_$ip"

# will not add label(agent_hostname) if true
omit_hostname = false

# s | ms
precision = "ms"

# global collect interval
interval = 15

# [global.labels]
# region = "shanghai"
# env = "localhost"

[writer_opt]
# default: 2000
batch = 2000
# channel(as queue) size
chan_size = 10000

[[writers]]
url = "https://monitserver.fastsdwan.com/prometheus/v1/write"

# Basic auth username
basic_auth_user = "CategrafUser"

# Basic auth password
basic_auth_pass = "ccaatteeggrraaff9527kubectl12#$"

# timeout settings, unit: ms
timeout = 5000
dial_timeout = 2500
max_idle_conns_per_host = 100

[http]
enable = false
address = ":9100"
print_access = false
run_mode = "release"
`,
	),
}

type configToml struct {
	Writers []configWriters `toml:"writers"`
}

type configWriters struct {
	Url           string `toml:"url"`
	basicAuthUser string `toml:"basic_auth_user"`
	basicAuthPass string `toml:"basic_auth_pass"`
}

func (t configToml) GetConfIndex(tomlStr string) ([]cIRecord, error) {
	var c configToml

	var cir []cIRecord
	_, err := toml.Decode(tomlStr, &c)
	if err != nil {
		return cir, err
	}

	for _, v := range configInfo.Indexes {
		switch v {
		case "writers.url":
			for _, i := range c.Writers {
				cir = append(cir, cIRecord{v, i.Url})
			}
		}
	}
	return cir, nil
}
