package ltwmodels

import (
	"github.com/didi/nightingale/v5/src/models"
	"time"
)

type TargetCtfItem struct {
	Id          int64  `json:"id" gorm:"primaryKey"`
	TargetId    int64  `json:"target_id"`
	Ip          string `json:"ip"`
	Hostname    string `json:"hostname"`
	Name        string `json:"name"`
	Status      string `json:"status"`
	LastToml    string `json:"last_toml"`
	CurrentToml string `json:"current_toml"`
	CreateAt    int64  `json:"create_at"`
	CreateBy    string `json:"create_by"`
	UpdateAt    int64  `json:"update_at"`
	UpdateBy    string `json:"update_by"`
}

type targetCtfItemStatus struct {
	ENABLED  string
	DISABLED string
}

var targetCtfItemStatusChoose = targetCtfItemStatus{
	ENABLED:  "ENABLED",
	DISABLED: "DISABLED",
}

func (tci *TargetCtfItem) TableName() string {
	return "ltw_target_ctf_item"
}

func TargetCtfItemGetByNameAndIp(ip, name string) (*TargetCtfItem, error) {
	session := models.DB().Where("name = ? and ip = ?", name, ip)
	var lst []*TargetCtfItem
	err := session.Find(&lst).Error
	if err != nil {
		return nil, err
	}

	if len(lst) == 0 {
		return nil, nil
	}
	return lst[0], nil

}

func (tci *TargetCtfItem) AddOrUpdate() error {
	r, err := TargetCtfItemGetByNameAndIp(tci.Name, tci.Ip)
	if err != nil {
		return err
	}
	now := time.Now().Unix()
	if r != nil {
		tci.UpdateAt = now
		return models.DB().Model(r).Select("*").Updates(tci).Error
		return nil
	}

	tci.CreateAt = now
	tci.UpdateAt = now
	return models.Insert(tci)
}
