package ctf

import (
	"github.com/BurntSushi/toml"
)

var mysqlInfo = CatConfInfo{
	Name:    "mysql",
	Indexes: []string{"instances.address"},
	Toml: string(`
# # collect interval
# interval = 15

[[instances]]
# address = "127.0.0.1:3306"
# username = "root"
# password = "1234"

# # set tls=custom to enable tls
# parameters = "tls=false"

# extra_status_metrics = true
# extra_innodb_metrics = false
# gather_processlist_processes_by_state = false
# gather_processlist_processes_by_user = false
# gather_schema_size = true
# gather_table_size = false
# gather_system_table_size = false
# gather_slave_status = true

# # timeout
# timeout_seconds = 3

# # interval = global.interval * interval_times
# interval_times = 1

# important! use global unique string to specify instance
# labels = { instance="n9e-10.2.3.4:3306" }

## Optional TLS Config
# use_tls = false
# tls_min_version = "1.2"
# tls_ca = "/etc/ctf/ca.pem"
# tls_cert = "/etc/ctf/cert.pem"
# tls_key = "/etc/ctf/key.pem"
## Use TLS but skip chain & host verification
# insecure_skip_verify = true

# [[instances.queries]]
# mesurement = "users"
# metric_fields = [ "total" ]
# label_fields = [ "service" ]
# # field_to_append = ""
# timeout = "3s"
# request = '''
# select 'n9e' as service, count(*) as total from n9e_v5.users
# '''
`,
	),
}

type mysqlToml struct {
	Interval  int             `toml:"interval"`
	Instances []mysqlInstance `toml:"instances"`
}

type mysqlInstance struct {
	Address string `toml:"address"`
}

func (t mysqlToml) GetConfIndex(tomlStr string) ([]cIRecord, error) {
	var c mysqlToml
	var cir []cIRecord

	_, err := toml.Decode(tomlStr, &c)
	if err != nil {
		return cir, err
	}

	for _, v := range mysqlInfo.Indexes {
		switch v {
		case "instances.address":
			for _, i := range c.Instances {
				cir = append(cir, cIRecord{v, i.Address})
			}
		}
	}
	return cir, nil
}
