package ltwmodels

import (
	"github.com/didi/nightingale/v5/src/models"
	"github.com/pkg/errors"
	"time"
)

type Voice struct {
	Id            int64  `json:"id" gorm:"primaryKey"`
	Result        string `json:"result"`
	AcceptTime    string `json:"accept_time"`
	CallFrom      string `json:"call_from"`
	Callid        string `json:"callid"`
	EndCalltime   string `json:"end_calltime"`
	Fee           string `json:"fee"`
	Mobile        string `json:"mobile"`
	Nationcode    string `json:"nationcode"`
	StartCalltime string `json:"start_calltime"`
	CreateAt      int64  `json:"create_at"`
}

func (v *Voice) TableName() string {
	return "ltw_voice"
}

func (v *Voice) Add() error {
	now := time.Now().Unix()
	v.CreateAt = now
	return models.Insert(v)
}

func VoiceGetByCallId(cid string) (*Voice, error) {
	session := models.DB().Where("callid=?", cid)

	var lst []*Voice
	err := session.Find(&lst).Error

	if len(lst) == 0 {
		return nil, errors.New("语音电话不存在！")
	}

	return lst[0], err
}
