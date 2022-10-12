package ctf

import (
	"fmt"
	"github.com/didi/nightingale/v5/src/ltwmodels"
	"path/filepath"
	"time"
)

type CatConfInfo struct {
	Name    string   `json:"name"`
	Indexes []string `json:"indexes"`
	Toml    string   `json:"toml"`
}

type confToml interface {
	GetConfIndex(tomlStr string) ([]cIRecord, error)
}

type cIRecord struct {
	index string
	Value string
}

func getTomlConfIndex(ct confToml, tomStr string) ([]cIRecord, error) {
	return ct.GetConfIndex(tomStr)
}

const ctfConfPath = "/opt/categraf/conf"

var CatConfArr = []CatConfInfo{
	configInfo,
	dnsQueryInfo,
	elasticsearchInfo,
	execInfo,
	httpResponseInfo,
	mongodbInfo,
	mysqlInfo,
	netResponseInfo,
	pingInfo,
	redisInfo,
	redisSentinelInfo,
	procstatInfo,
}

func CreateQueryIndexes(hc ltwmodels.HostCtfConf) error {
	var cir []cIRecord
	var err error
	switch hc.Name {
	case "config":
		cir, err = getTomlConfIndex(new(configToml), hc.LocalToml)
	case "dns_query":
		cir, err = getTomlConfIndex(new(dnsQueryToml), hc.LocalToml)
	case "elasticsearch":
		cir, err = getTomlConfIndex(new(elasticsearchToml), hc.LocalToml)
	case "exec":
		cir, err = getTomlConfIndex(new(execToml), hc.LocalToml)
	case "http_response":
		cir, err = getTomlConfIndex(new(httpResponseToml), hc.LocalToml)
	case "mongodb":
		cir, err = getTomlConfIndex(new(mongodbToml), hc.LocalToml)
	case "mysql":
		cir, err = getTomlConfIndex(new(mysqlToml), hc.LocalToml)
	case "net_response":
		cir, err = getTomlConfIndex(new(netResponseToml), hc.LocalToml)
	case "ping":
		cir, err = getTomlConfIndex(new(pingToml), hc.LocalToml)
	case "redis":
		cir, err = getTomlConfIndex(new(redisToml), hc.LocalToml)
	case "redis_sentinel":
		cir, err = getTomlConfIndex(new(redisSentinelToml), hc.LocalToml)

	case "procstat":
		cir, err = getTomlConfIndex(new(procstatToml), hc.LocalToml)
	default:
		fmt.Println("格式不支持，请联系管理员")
	}
	if err != nil {
		return err
	}

	for _, v := range cir {
		hci := ltwmodels.HostCtfConfIndex{
			HostCtfConfId: hc.Id,
			Name:          hc.Name,
			Index:         v.index,
			Value:         v.Value,
		}
		hci.Add()
	}
	return nil
}

func GetConfigFilePathByName(name string) string {
	dirName := "input." + name
	fileName := name + ".toml"
	if name == "config" {
		return filepath.Join(ctfConfPath, fileName)
	}
	return filepath.Join(ctfConfPath, dirName, fileName)
}

func GetConfigDirPathByName(name string) string {
	dirName := "input." + name
	if name == "config" {
		return ctfConfPath
	}
	return filepath.Join(ctfConfPath, dirName)
}

func GetBatConfigDirPathByName(name string) string {
	dirName := "input." + name
	timeTag := time.Now().Format("20060102030405")
	return filepath.Join(ctfConfPath, "bat."+timeTag+"."+dirName)
}
