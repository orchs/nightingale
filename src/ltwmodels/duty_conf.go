package ltwmodels

import (
	"github.com/didi/nightingale/v5/src/models"
	"github.com/pkg/errors"
	"github.com/toolkits/pkg/str"
	"time"
)

type DutyConf struct {
	Id       int64  `json:"id" gorm:"primaryKey"`
	GroupId  int64  `json:"group_id"`
	Name     string `json:"name"`
	StartAt  string `json:"start_at"`
	EndAt    string `json:"end_at"`
	Priority int64  `json:"priority"`
	Using    int    `json:"using"`
	CreateAt int64  `json:"create_at"`
	CreateBy string `json:"create_by"`
	UpdateAt int64  `json:"update_at"`
	UpdateBy string `json:"update_by"`
}

func (dc *DutyConf) TableName() string {
	return "ltw_duty_conf"
}

func (dc *DutyConf) Verify() error {
	if dc.Name == "" {
		return errors.New("班次名称不能为空！")
	}

	if str.Dangerous(dc.Name) {
		return errors.New("班次名称不合规！")
	}

	if dc.StartAt == "" {
		return errors.New("班次开始时间不能为空！")
	}

	if dc.EndAt == "" {
		return errors.New("班次结束时间不能为空！")
	}

	return nil
}

func (dc *DutyConf) Add() error {
	if err := dc.Verify(); err != nil {
		return err
	}

	exists, err := DutyConfExists(0, dc.GroupId, dc.Name)
	if err != nil {
		return err
	}

	if exists {
		return errors.New("该用户组已存在此班次，请修改班次名称！")
	}

	now := time.Now().Unix()
	dc.CreateAt = now
	dc.UpdateAt = now
	dc.Using = 1

	return models.Insert(dc)
}

func (dc *DutyConf) Update(dcf DutyConf) error {
	if dc.Name != dcf.Name {
		exists, err := DutyConfExists(dc.Id, dc.GroupId, dc.Name)
		if err != nil {
			return err
		}

		if exists {
			return errors.New("该用户组已存在此班次，请修改班次名称！")
		}
	}

	dcf.Id = dc.Id
	dcf.GroupId = dc.GroupId
	dcf.CreateAt = dc.CreateAt
	dcf.CreateBy = dc.CreateBy
	dcf.UpdateAt = time.Now().Unix()
	err := dcf.Verify()
	if err != nil {
		return err
	}
	return models.DB().Model(dc).Select("*").Updates(dcf).Error
}

func (dc *DutyConf) Delete() error {
	dc.Using = 0
	dc.UpdateAt = time.Now().Unix()

	return models.DB().Model(dc).Select("*").Updates(dc).Error
}

func DutyConfExists(id, groupId int64, name string) (bool, error) {
	session := models.DB().Where("id <> ? and group_id = ? and name = ?", id, groupId, name)

	var lst []DutyConf
	err := session.Find(&lst).Error
	if err != nil {
		return false, err
	}
	if len(lst) == 0 {
		return false, nil
	}

	return true, nil
}

func DutyConfGetByGid(gid int64) ([]DutyConf, error) {
	session := models.DB().Where("`using`=1 and group_id=?", gid).Order("priority")

	var lst []DutyConf
	err := session.Find(&lst).Error

	return lst, err
}

func DutyConfGetById(id int64) (*DutyConf, error) {
	session := models.DB().Where("id=?", id).Order("priority")

	var lst []*DutyConf
	err := session.Find(&lst).Error

	if len(lst) == 0 {
		return nil, errors.New("要修改的班次不存在！")
	}
	return lst[0], err
}

func DutyConfGet(where string, args ...interface{}) (*DutyConf, error) {
	var lst []*DutyConf
	err := models.DB().Where(where, args...).Find(&lst).Error
	if err != nil {
		return nil, err
	}

	if len(lst) == 0 {
		return nil, nil
	}

	return lst[0], nil
}
