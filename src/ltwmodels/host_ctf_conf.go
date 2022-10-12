package ltwmodels

import (
	"context"
	"encoding/json"
	"github.com/didi/nightingale/v5/src/models"
	"github.com/didi/nightingale/v5/src/storage"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/toolkits/pkg/ginx"
	"time"
)

type HostCtfConf struct {
	Id         int64  `json:"id" gorm:"primaryKey"`
	HostCtfId  int64  `json:"host_ctf_id"`
	Ip         string `json:"ip"`
	Name       string `json:"name"`
	Status     string `json:"status"`
	LocalToml  string `json:"local_toml"`
	RemoteToml string `json:"remote_toml"`
	CreateAt   int64  `json:"create_at"`
	CreateBy   string `json:"create_by"`
	UpdateAt   int64  `json:"update_at"`
	UpdateBy   string `json:"update_by"`
}

type hostCtfConfStatus struct {
	UNINSTALLED string
	INSTALLED   string
	CONFLICTING string
	ERROR       string
}

var HCStatus = hostCtfConfStatus{
	UNINSTALLED: "UNINSTALLED",
	INSTALLED:   "INSTALLED",
	CONFLICTING: "CONFLICTING",
	ERROR:       "ERROR",
}

func (hc *HostCtfConf) TableName() string {
	return "ltw_host_ctf_conf"
}

func (hc *HostCtfConf) Verify() error {
	if hc.Ip == "" {
		return errors.New("HostCtfConf filed Ip is blank")
	}
	return nil
}

func hostCtfConfExists(name, ip string) (bool, error) {
	session := models.DB().Where("name = ? and ip = ?", name, ip)

	var lst []HostCtfConf
	err := session.Find(&lst).Error
	if err != nil {
		return false, err
	}

	if len(lst) > 0 {
		return true, nil
	}

	return false, nil
}

func HostCtfConfGetByIp(ip string) ([]*HostCtfConf, error) {
	return hostCtfConfGet("ip=?", ip)
}

func HostCtfConfGet(where string, args ...interface{}) (HostCtfConf, error) {
	var lst []HostCtfConf
	var res HostCtfConf
	err := models.DB().Where(where, args...).Find(&lst).Error
	if err != nil {
		return res, err
	}

	if len(lst) == 0 {
		return res, nil
	}
	res = lst[0]
	return res, nil
}

func HostCtfConfGetByIpName(ip, name string) (HostCtfConf, error) {
	return HostCtfConfGet("ip=? and name=?", ip, name)
}

func HostCtfConfGetById(id int64) (HostCtfConf, error) {
	return HostCtfConfGet("id=?", id)
}
func (hc *HostCtfConf) Add() (HostCtfConf, error) {
	var newHc HostCtfConf
	if err := hc.Verify(); err != nil {
		return newHc, err
	}
	exists, err := hostCtfConfExists(hc.Name, hc.Ip)
	if err != nil {
		return newHc, err
	}

	if exists {
		return newHc, errors.New("Host CTF already exists")
	}

	now := time.Now().Unix()
	hc.CreateAt = now
	hc.UpdateAt = now
	res, err := models.Insert2Obj(hc)
	ginx.Dangerous(err)
	m, err := json.Marshal(res)
	jsonRec := json.Unmarshal(m, &newHc)
	if jsonRec != nil {
		return newHc, jsonRec
	}
	return newHc, nil
}

func (hc *HostCtfConf) AddHostCtfConfLog(taskStatus, msg, lastToml string) error {
	ctx := context.Background()
	host, err := storage.Redis.HGetAll(ctx, "host_"+hc.Ip).Result()
	if err != nil {
		return err
	}
	h := CmdbHostInfo{}
	err = mapstructure.Decode(host, &h)
	if err != nil {
		return err
	}

	hcl := HostCtfConfLogs{
		Hostname:      h.HostName,
		Ip:            hc.Ip,
		HostCtfConfId: hc.Id,
		Name:          hc.Name,
		Status:        taskStatus,
		Message:       msg,
		LastToml:      lastToml,
		CurrentToml:   hc.RemoteToml,
		CreateBy:      hc.CreateBy,
		UpdateBy:      hc.UpdateBy,
	}
	if err := hcl.Add(); err != nil {
		return err
	}
	return nil
}

func hostCtfConfGet(where string, args ...interface{}) ([]*HostCtfConf, error) {
	var lst []*HostCtfConf
	err := models.DB().Where(where, args...).Find(&lst).Error
	if err != nil {
		return nil, err
	}

	return lst, nil
}

func (hc *HostCtfConf) Update(newMf HostCtfConf) error {
	newMf.Id = hc.Id
	newMf.HostCtfId = hc.HostCtfId
	newMf.Name = hc.Name
	newMf.CreateAt = hc.CreateAt
	newMf.CreateBy = hc.CreateBy
	newMf.UpdateAt = time.Now().Unix()

	err := newMf.Verify()
	if err != nil {
		return err
	}
	return models.DB().Model(hc).Select("*").Updates(newMf).Error
}

func (hc *HostCtfConf) Del() error {
	return models.DB().Where("id=?", hc.Id).Delete(&HostCtfConf{}).Error
}

func (hc *HostCtfConf) DelIndexById() error {
	return models.DB().Where("host_ctf_conf_id=?", hc.Id).Delete(&HostCtfConfIndex{}).Error
}

func GetIpsByName(name string) []string {
	// 通过监控项id查询主机ip列表
	session := models.DB().Model(&HostCtfConf{}).Select("distinct ip").Where("name = ?", name)

	var ips []string
	err := session.Find(&ips).Error
	ginx.Dangerous(err)

	return ips
}

func GetIpsByIndexValue(index, value string) []string {
	// 通过索引表中的值查询主机ip列表
	session := models.DB().Model(&HostCtfConf{}).Select("distinct ip").Joins(
		"LEFT JOIN ltw_host_ctf_conf_index ON ltw_host_ctf_conf_index.host_categraf_id = ltw_host_ctf_conf.id",
	).Where("ltw_host_ctf_conf_index.index = ? and value = ?", index, value)

	var ips []string
	err := session.Find(&ips).Error
	ginx.Dangerous(err)

	return ips

}

func GetHostCtfConfByIp(ip string) ([]HostCtfConf, error) {
	session := models.DB().Model(&HostCtfConf{}).Where("ip = ?", ip)
	var res []HostCtfConf
	err := session.Find(&res).Error
	return res, err
}
