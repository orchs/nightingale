package ltwmodels

import (
	"github.com/didi/nightingale/v5/src/models"
	"github.com/pkg/errors"
	"time"
)

type HostCtfConfIndex struct {
	Id            int64  `json:"id" gorm:"primaryKey"`
	HostCtfConfId int64  `json:"host_ctf_conf_id"`
	Name          string `json:"name"`
	Index         string `json:"index"`
	Value         string `json:"value"`
	CreateAt      int64  `json:"create_at"`
	CreateBy      string `json:"create_by"`
	UpdateAt      int64  `json:"update_at"`
	UpdateBy      string `json:"update_by"`
}

func (hci *HostCtfConfIndex) TableName() string {
	return "ltw_host_ctf_conf_index"
}

func (hci *HostCtfConfIndex) Add() error {
	now := time.Now().Unix()
	hci.CreateAt = now
	hci.UpdateAt = now

	return models.Insert(hci)
}

func (hci *HostCtfConfIndex) Del() error {
	return models.DB().Where("id=?", hci.Id).Delete(&HostCtfConfIndex{}).Error
}

func GetHostCtfIdByIndex(index string) ([]HostCtfConfIndex, error) {

	session := models.DB().Where("index = ?", index)

	var res []HostCtfConfIndex
	err := session.Find(&res).Error
	if err != nil {
		return res, errors.WithMessage(
			err, "failed to query ctf index",
		)
	}

	return res, nil
}
