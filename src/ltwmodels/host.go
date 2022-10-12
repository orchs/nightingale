package ltwmodels

type CmdbHostInfo struct {
	HostName  string   `json:"hostname"`
	Ip        string   `json:"ip"`
	Port      string   `json:"port"`
	PrivateIp string   `json:"private_ip"`
	Status    string   `json:"status"`
	AdminUser string   `json:"admin_user"`
	Rsa       string   `json:"rsa"`
	Passwd    string   `json:"passwd"`
	Actions   []string `json:"actions"`
	User      string   `json:"user"`
}
