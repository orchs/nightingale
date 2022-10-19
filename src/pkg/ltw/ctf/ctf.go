package ctf

type ItemInfo struct {
	Name   string `json:"name"`
	Status bool   `json:"status"`
	Toml   string `json:"toml"`
}

var Items = []ItemInfo{
	{
		Name:   "config",
		Status: false,
		Toml:   configInfo.Toml,
	},
	{
		Name:   "dns_query",
		Status: false,
		Toml:   dnsQueryInfo.Toml,
	},
	{
		Name:   "elasticsearch",
		Status: false,
		Toml:   elasticsearchInfo.Toml,
	},
	{
		Name:   "exec",
		Status: false,
		Toml:   execInfo.Toml,
	},
	{
		Name:   "http_response",
		Status: false,
		Toml:   httpResponseInfo.Toml,
	},
	{
		Name:   "mongodb",
		Status: false,
		Toml:   mongodbInfo.Toml,
	},
	{
		Name:   "net_response",
		Status: false,
		Toml:   netResponseInfo.Toml,
	},
	{
		Name:   "nginx_upstream_check",
		Status: false,
		Toml:   nginxUpstreamCheckInfo.Toml,
	},
	{
		Name:   "ping",
		Status: false,
		Toml:   pingInfo.Toml,
	},
	{
		Name:   "procstat",
		Status: false,
		Toml:   procstatInfo.Toml,
	},
	{
		Name:   "redis",
		Status: false,
		Toml:   redisInfo.Toml,
	},
	{
		Name:   "redis_sentinel",
		Status: false,
		Toml:   redisSentinelInfo.Toml,
	},
}
