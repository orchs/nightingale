// ltw

package ltwmodels

import (
	"gorm.io/gorm"
	"strings"
	"time"
)

import (
	"github.com/didi/nightingale/v5/src/models"
)

type CtfLogs struct {
	Id       int64  `json:"id" gorm:"primaryKey"`
	Ip       string `json:"ip"`
	Hostname string `json:"hostname"`
	Action   string `json:"action"`
	Status   string `json:"status"`
	StandOut string `json:"stand_out"`
	CreateAt int64  `json:"create_at"`
	CreateBy string `json:"create_by"`
}

func (hcl *CtfLogs) TableName() string {
	return "ltw_ctf_logs"
}

type CtfLogStatus struct {
	SUCCEED string
	FAILED  string
}

var CtfLogStatusChoose = CtfLogStatus{
	SUCCEED: "succeed",
	FAILED:  "failed",
}

type CtfLogAction struct {
	INSTALL   string
	UNINSTALL string
}

var CtfLogActionChoose = CtfLogAction{
	INSTALL:   "install",
	UNINSTALL: "uninstall",
}

func (hcl *CtfLogs) Add() error {
	now := time.Now().Unix()
	hcl.CreateAt = now
	return models.Insert(hcl)
}

func CtfLogsTotal(ip, name, status string) (int64, error) {
	return models.Count(buildLogWhere(ip, name, status))
}

func GetCtfLogs(ip, name, status string, limit, offset int) ([]*CtfLogs, error) {
	var lst []*CtfLogs
	err := buildCtfLogsWhere(ip, name, status).Order("id desc").Limit(limit).Offset(offset).Find(&lst).Error
	return lst, err
}

func buildCtfLogsWhere(query, status, action string) *gorm.DB {
	session := models.DB().Model(&CtfLogs{})

	if query != "" {
		arr := strings.Fields(query)
		for i := 0; i < len(arr); i++ {
			q := "%" + arr[i] + "%"
			session = session.Where("ip like ? or hostname like ?", q, q, q)
		}
	}

	if status != "" {
		session = session.Where("status = ?", status)
	}

	if action != "" {
		session = session.Where("action = ?", action)
	}

	return session
}
