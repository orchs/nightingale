package ctf

import (
	"github.com/BurntSushi/toml"
)

var influxdbInfo = CatConfInfo{
	Name:    "influxdb",
	Indexes: []string{"instances.urls"},
	Toml: string(`
# # collect interval
# interval = 15

[[instances]]

urls = [
    "http://81.70.0.214:8088/debug/vars"
]
## Username and password to send using HTTP Basic Authentication.
# username = ""
# password = ""

## Optional TLS Config
# tls_ca = "/etc/telegraf/ca.pem"
# tls_cert = "/etc/telegraf/cert.pem"
# tls_key = "/etc/telegraf/key.pem"
## Use TLS but skip chain & host verification
# insecure_skip_verify = false

## http request & header timeout


# # interval = global.interval * interval_times
# interval_times = 1

# important! use global unique string to specify instance
# labels = { instance="n9e-10.2.3.4:6379" }

## Optional TLS Config
# use_tls = false
# tls_min_version = "1.2"
# tls_ca = "/etc/categraf/ca.pem"
# tls_cert = "/etc/categraf/cert.pem"
# tls_key = "/etc/categraf/key.pem"
## Use TLS but skip chain & host verification
# insecure_skip_verify = true
`,
	),
}

type influxdbToml struct {
	Interval  int                `toml:"interval"`
	Instances []influxdbInstance `toml:"instances"`
}

type influxdbInstance struct {
	Urls []string `toml:"urls"`
}

func (t influxdbToml) GetConfIndex(tomlStr string) ([]cIRecord, error) {
	var c influxdbToml
	var cir []cIRecord

	_, err := toml.Decode(tomlStr, &c)
	if err != nil {
		return cir, err
	}

	for _, v := range influxdbInfo.Indexes {
		switch v {
		case "instances.address":
			for _, i := range c.Instances {
				for _, t := range i.Urls {
					cir = append(cir, cIRecord{v, t})
				}
			}
		}
	}
	return cir, nil
}
