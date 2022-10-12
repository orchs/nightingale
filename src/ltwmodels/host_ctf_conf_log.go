package ltwmodels

import (
	"context"
	"github.com/didi/nightingale/v5/src/models"
	"github.com/didi/nightingale/v5/src/storage"
	"github.com/mitchellh/mapstructure"
	"gorm.io/gorm"
	"time"
)

type HostCtfConfLogs struct {
	Id            int64  `json:"id" gorm:"primaryKey"`
	Ip            string `json:"ip"`
	Hostname      string `json:"hostname"`
	HostCtfConfId int64  `json:"host_ctf_conf_id"`
	Name          string `json:"name"`
	Status        string `json:"status"`
	Message       string `json:"message"`
	LastToml      string `json:"last_toml"`
	CurrentToml   string `json:"current_toml"`
	StandOut      string `json:"stand_out"`
	CreateAt      int64  `json:"create_at"`
	CreateBy      string `json:"create_by"`
	UpdateAt      int64  `json:"update_at"`
	UpdateBy      string `json:"update_by"`
}
type LogStatus struct {
	SUCCEED string
	FAILED  string
}

var CtfConfLogStatus = LogStatus{
	SUCCEED: "succeed",
	FAILED:  "failed",
}

func (hcl *HostCtfConfLogs) TableName() string {
	return "ltw_host_ctf_conf_log"
}

func (hcl *HostCtfConfLogs) Add() error {
	now := time.Now().Unix()
	hcl.CreateAt = now
	hcl.UpdateAt = now

	return models.Insert(hcl)
}

func (hcl *HostCtfConfLogs) Del() error {
	return models.DB().Where("id=?", hcl.Id).Delete(&HostCtfConfLogs{}).Error
}

func buildLogWhere(ip, name, status string) *gorm.DB {
	session := models.DB().Model(&HostCtfConfLogs{})
	if ip != "" {
		session = session.Where("ip = ?", ip)
	}
	if name != "" {
		session = session.Where("name = ?", name)
	}
	if status != "" {
		session = session.Where("status = ?", status)
	}
	return session
}

func HostCtfLogsTotal(ip, name, status string) (int64, error) {
	return models.Count(buildLogWhere(ip, name, status))
}

func GetHostCtfLogs(ip, name, status string, limit, offset int) ([]*HostCtfConfLogs, error) {
	var lst []*HostCtfConfLogs
	err := buildLogWhere(ip, name, status).Order("id desc").Limit(limit).Offset(offset).Find(&lst).Error
	return lst, err
}

func AddHostCtfConfLog(pid int64, ip, name, status, msg, lastToml, currentToml, stdOut, username string) error {
	ctx := context.Background()
	host, err := storage.Redis.HGetAll(ctx, "host_"+ip).Result()
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
		Ip:            ip,
		HostCtfConfId: pid,
		Name:          name,
		Status:        status,
		Message:       msg,
		LastToml:      lastToml,
		CurrentToml:   currentToml,
		StandOut:      stdOut,
		CreateBy:      username,
		UpdateBy:      username,
	}
	if err := hcl.Add(); err != nil {
		return err
	}
	return nil
}
