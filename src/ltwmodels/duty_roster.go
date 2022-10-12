package ltwmodels

import (
	"fmt"
	"github.com/didi/nightingale/v5/src/models"
	"github.com/pkg/errors"
	"strconv"
	"time"
)

type DutyRoster struct {
	Id         int64  `json:"id" gorm:"primaryKey"`
	GroupId    int64  `json:"group_id"`
	DutyConfId int64  `json:"duty_conf_id"`
	UserId     int64  `json:"user_id"`
	StartAt    int64  `json:"start_at"`
	EndAt      int64  `json:"end_at"`
	DutyDate   int64  `json:"duty_date"`
	CreateAt   int64  `json:"create_at"`
	CreateBy   string `json:"create_by"`
	UpdateAt   int64  `json:"update_at"`
	UpdateBy   string `json:"update_by"`
}

const DateTimeFmt = "20060102 15:04:05"

func (dr *DutyRoster) TableName() string {
	return "ltw_duty_roster"
}

func (dr *DutyRoster) Verify() error {
	if dr.UserId == 0 {
		return errors.New("用户id不能为空！")
	}

	if dr.StartAt >= dr.EndAt {
		return errors.New("值班结束时间不能小于开始时间！")
	}

	return nil
}

func DutyRosterExists(dutyConfId, dutyDate, userId int64) (bool, error) {
	session := models.DB().Where("duty_conf_id=?  and duty_date=? and user_id=?", dutyConfId, dutyDate, userId)

	var lst []DutyRoster
	err := session.Find(&lst).Error
	if err != nil {
		return false, err
	}
	if len(lst) == 0 {
		return false, nil
	}

	return true, nil
}

func GetDutyRangeTimeUnix(ymd int64, s, e string) (int64, int64) {
	local, _ := time.LoadLocation("Asia/Shanghai")

	sts := strconv.FormatInt(ymd, 10) + " " + s
	st, err := time.ParseInLocation(DateTimeFmt, sts, local)
	if err != nil {
		fmt.Println(err.Error())
	}

	ets := strconv.FormatInt(ymd, 10) + " " + e
	et, err := time.ParseInLocation(DateTimeFmt, ets, local)
	if err != nil {
		fmt.Println(err.Error())
	}

	if st.Unix() >= et.Unix() {
		et = et.AddDate(0, 0, 1)
	}

	return st.Unix(), et.Unix()
}

func (dr *DutyRoster) Add() error {
	exists, err := DutyRosterExists(dr.DutyConfId, dr.DutyDate, dr.UserId)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	dc, err := DutyConfGet("id=?", dr.DutyConfId)
	if err != nil {
		return err
	}

	dr.StartAt, dr.EndAt = GetDutyRangeTimeUnix(dr.DutyDate, dc.StartAt, dc.EndAt)
	if err := dr.Verify(); err != nil {
		return err
	}

	now := time.Now().Unix()
	dr.CreateAt = now
	dr.UpdateAt = now

	return models.Insert(dr)
}

func DelByDutyAndDate(did, dd int64) error {
	return models.DB().Where(
		"duty_conf_id=? and duty_date=?", did, dd,
	).Delete(&DutyRoster{}).Error

}

func DutyRosterRecordGets(gid, sd, ed int64) ([]*DutyRoster, error) {
	return DutyRosterGets("group_id=? and duty_date>=? and duty_date<=?", gid, sd, ed)
}

func DutyRosterWatchkeeperGets(gid, t int64) ([]int64, error) {
	session := models.DB().Model(&DutyRoster{}).Select("user_id").Joins(
		"INNER JOIN ltw_duty_conf ON ltw_duty_conf.id = ltw_duty_roster.duty_conf_id",
	).Where("ltw_duty_roster.group_id=? and ltw_duty_roster.start_at<=? and ltw_duty_roster.end_at>=?", gid, t, t).Order("priority desc, ltw_duty_roster.id asc")

	var ids []int64
	err := session.Find(&ids).Error

	return ids, err
}

func DutyRosterGets(where string, args ...interface{}) ([]*DutyRoster, error) {
	var lst []*DutyRoster
	err := models.DB().Where(where, args...).Find(&lst).Error
	if err != nil {
		return nil, err
	}

	if len(lst) == 0 {
		return nil, nil
	}

	return lst, nil
}
