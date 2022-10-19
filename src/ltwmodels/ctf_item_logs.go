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

type CtfItemLogs struct {
	Id       int64  `json:"id" gorm:"primaryKey"`
	Ip       string `json:"ip"`
	Hostname string `json:"hostname"`
	Ident    string `json:"ident"`
	Action   string `json:"action"`
	Status   string `json:"status"`
	Toml     string `json:"toml"`
	StandOut string `json:"stand_out"`
	CreateAt int64  `json:"create_at"`
	CreateBy string `json:"create_by"`
}

func (hcl *CtfItemLogs) TableName() string {
	return "ltw_ctf_item_logs"
}

type CtfItemLogStatus struct {
	SUCCEED string
	FAILED  string
}

var CtfItemLogStatusChoose = CtfItemLogStatus{
	SUCCEED: "succeed",
	FAILED:  "failed",
}

type CtfItemLogAction struct {
	ADD    string
	UPDATE string
}

var CtfItemLogActionChoose = CtfItemLogAction{
	ADD:    "add",
	UPDATE: "update",
}

func (hcl *CtfItemLogs) Add() error {
	now := time.Now().Unix()
	hcl.CreateAt = now
	return models.Insert(hcl)
}

func CtfItemLogsTotal(ip, name, status string) (int64, error) {
	return models.Count(buildLogWhere(ip, name, status))
}

func GetCtfItemLogs(ip, name, status string, limit, offset int) ([]*CtfItemLogs, error) {
	var lst []*CtfItemLogs
	err := buildCtfItemLogsWhere(ip, name, status).Order("id desc").Limit(limit).Offset(offset).Find(&lst).Error
	return lst, err
}

func buildCtfItemLogsWhere(query, status, action string) *gorm.DB {
	session := models.DB().Model(&CtfItemLogs{})

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
