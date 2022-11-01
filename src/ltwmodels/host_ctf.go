package ltwmodels

import (
	"github.com/didi/nightingale/v5/src/models"
	"github.com/pkg/errors"
	"github.com/toolkits/pkg/str"
	"time"
)

type HostCtf struct {
	Id       int64  `json:"id" gorm:"primaryKey"`
	Ip       string `json:"ip"`
	Hostname string `json:"hostname"`
	Status   string `json:"status"`
	Version  string `json:"version"`
	Actions  string `json:"actions"`
	CreateAt int64  `json:"create_at"`
	CreateBy string `json:"create_by"`
	UpdateAt int64  `json:"update_at"`
	UpdateBy string `json:"update_by"`
}

func (hc *HostCtf) TableName() string {
	return "ltw_host_ctf"
}

type hostCtfStatus struct {
	UNINSTALLED string
	ENABLED     string
	DISABLED    string
}

var HostCtfStatus = hostCtfStatus{
	ENABLED:     "ENABLED",
	DISABLED:    "DISABLED",
	UNINSTALLED: "UNINSTALLED",
}

func (hc *HostCtf) Verify() error {
	if hc.Ip == "" {
		return errors.New("HostCtf filed Ip is blank")
	}

	if str.Dangerous(hc.Hostname) {
		return errors.New("HostCtf has invalid characters")
	}

	return nil
}

func HostCtfGet(where string, args ...interface{}) (HostCtf, error) {
	var lst []HostCtf
	var res HostCtf
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

func GetHostCtfByConf(name, index, value string) ([]HostCtf, error) {
	// 通过索引表中的值查询主机ip列表
	session := models.DB().Model(&HostCtf{}).Select("distinct ltw_host_ctf.ip, ltw_host_ctf.status, ltw_host_ctf.version").Joins(
		"LEFT JOIN ltw_host_ctf_conf ON ltw_host_ctf_conf.ip = ltw_host_ctf.ip",
	).Joins(
		"LEFT JOIN ltw_host_ctf_conf_index ON ltw_host_ctf_conf_index.host_ctf_conf_id = ltw_host_ctf_conf.id",
	).Where("ltw_host_ctf_conf.name = ? ", name)

	if index != "" {
		session = session.Where("ltw_host_ctf_conf_index.index = ? ", index)
	}

	if value != "" {
		session = session.Where("ltw_host_ctf_conf_index.value = ?", value)
	}

	var res []HostCtf
	err := session.Find(&res).Error
	return res, err
}

func GetHostCtfByIps(ips []string) ([]HostCtf, error) {
	// 通过主机ip查找记录
	session := models.DB().Model(&HostCtf{}).Where("ip in ?", ips)
	var res []HostCtf
	err := session.Find(&res).Error
	return res, err
}

func GetHostCtfByIp(ip string) (HostCtf, error) {
	return HostCtfGet("ip=?", ip)
}

func (hc *HostCtf) Save() error {
	now := time.Now().Unix()
	hc.CreateAt = now
	hc.UpdateAt = now
	var oldHc HostCtf
	var err error
	if oldHc, err = GetHostCtfByIp(hc.Ip); err != nil {
		return err
	} else if oldHc.Ip == "" {
		return models.Insert(hc)
	} else {
		hc.Id = oldHc.Id
		hc.CreateAt = oldHc.CreateAt
		hc.CreateBy = oldHc.CreateBy
		hc.UpdateAt = time.Now().Unix()
		return models.DB().Model(hc).Select("*").Updates(hc).Error
	}
	return nil
}
