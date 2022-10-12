package ctf

import (
	"github.com/BurntSushi/toml"
)

var prometheusInfo = CatConfInfo{
	Name:    "prometheus",
	Indexes: []string{"instances.urls"},
	Toml: string(`
# # collect interval
# interval = 15

[[instances]]
urls = [
#     "http://localhost:9104/metrics"
]

url_label_key = "instance"
url_label_value = "{{.Host}}"

## Scrape Services available in Consul Catalog
# [instances.consul]
#   enabled = false
#   agent = "http://localhost:8500"
#   query_interval = "5m"

#   [[instances.consul.query]]
#     name = "a service name"
#     tag = "a service tag"
#     url = 'http://{{if ne .ServiceAddress ""}}{{.ServiceAddress}}{{else}}{{.Address}}{{end}}:{{.ServicePort}}/{{with .ServiceMeta.metrics_path}}{{.}}{{else}}metrics{{end}}'
#     [instances.consul.query.tags]
#       host = "{{.Node}}"

# bearer_token_string = ""

# e.g. /run/secrets/kubernetes.io/serviceaccount/token
# bearer_token_file = ""

# # basic auth
# username = ""
# password = ""

# headers = ["X-From", "categraf"]

# # interval = global.interval * interval_times
# interval_times = 1

# labels = {}

# support glob
# ignore_metrics = [ "go_*" ]

# support glob
# ignore_label_keys = []

# timeout for every url
# timeout = "3s"

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

type prometheusToml struct {
	Interval  int                  `toml:"interval"`
	Instances []prometheusInstance `toml:"instances"`
}

type prometheusInstance struct {
	Urls []string `toml:"urls"`
}

func (t prometheusToml) GetConfIndex(tomlStr string) ([]cIRecord, error) {
	var c prometheusToml
	var cir []cIRecord

	_, err := toml.Decode(tomlStr, &c)
	if err != nil {
		return cir, err
	}

	for _, v := range prometheusInfo.Indexes {
		switch v {
		case "instances.urls":
			for _, i := range c.Instances {
				for _, t := range i.Urls {
					cir = append(cir, cIRecord{v, t})
				}
			}
		}
	}
	return cir, err
}
