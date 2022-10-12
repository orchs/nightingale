package ctf

import (
	"github.com/BurntSushi/toml"
)

var elasticsearchInfo = CatConfInfo{
	Name:    "elasticsearch",
	Indexes: []string{"instances.servers"},
	Toml: string(`
interval = 15
[[instances]]
interval_times = 1
labels = { cluster="orch_ope-es" }
servers = ["http://127.0.0.1:9200"]
http_timeout = "5s"
local = true
cluster_health = true
cluster_health_level = "cluster"
cluster_stats = true
indices_include = ["ipcache_*"]
indices_level = ""
node_stats = ["jvm", "breaker", "process", "os", "fs", "indices", "http", "thread_pool", "transport"]
username = ""
password = ""
use_tls = false
tls_ca = "/etc/ctf/ca.pem"
tls_cert = "/etc/ctf/cert.pem"
tls_key = "/etc/ctf/key.pem"
# insecure_skip_verify = true
num_most_recent_indices = 10
`,
	),
}

type elasticsearchToml struct {
	Interval  int                     `toml:"interval"`
	Instances []elasticsearchInstance `toml:"instances"`
}

type elasticsearchInstance struct {
	Servers []string `toml:"servers"`
}

func (t elasticsearchToml) GetConfIndex(tomlStr string) ([]cIRecord, error) {
	var c elasticsearchToml
	var cir []cIRecord

	_, err := toml.Decode(tomlStr, &c)
	if err != nil {
		return cir, err
	}

	for _, v := range elasticsearchInfo.Indexes {
		switch v {
		case "instances.servers":
			for _, i := range c.Instances {
				for _, t := range i.Servers {
					cir = append(cir, cIRecord{v, t})
				}
			}
		}
	}
	return cir, nil
}
