package ltw

import (
	"github.com/didi/nightingale/v5/src/ltwmodels"
	"github.com/didi/nightingale/v5/src/models"
	"github.com/gin-gonic/gin"
	"github.com/toolkits/pkg/ginx"
	"strconv"
	"time"
)

type operatorForm struct {
	UserIds    []int64 `json:"user_ids"`
	DutyDate   int64   `json:"duty_date"`
	DutyConfId int64   `json:"duty_conf_id"`
}
type person struct {
	UserId   int64  `json:"user_id"`
	Username string `json:"username"`
}
type watchkeeper struct {
	DutyDate int64    `json:"duty_date"`
	Persons  []person `json:"persons"`
}

type weekDutyRoster struct {
	DutyConfId       int64       `json:"duty_conf_id"`
	DutyConfName     string      `json:"duty_conf_name"`
	DutyIntervalTime string      `json:"duty_range_time"`
	Monday           watchkeeper `json:"monday"`
	Tuesday          watchkeeper `json:"tuesday"`
	Wednesday        watchkeeper `json:"wednesday"`
	Thursday         watchkeeper `json:"thursday"`
	Friday           watchkeeper `json:"friday"`
	Saturday         watchkeeper `json:"saturday"`
	Sunday           watchkeeper `json:"sunday"`
}

func GetUsersByGid(gid int64) ([]models.User, error) {
	ids, err := models.MemberIds(gid)
	if err != nil {
		return nil, err
	}
	users, err := models.UserGetsByIds(ids)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func GetUserMapByGid(gid int64) (map[int64]string, error) {
	users, err := GetUsersByGid(gid)
	if err != nil {
		return nil, err
	}
	m := make(map[int64]string)
	for _, u := range users {
		m[u.Id] = u.Username
	}
	return m, nil
}

func WatchkeeperGets(c *gin.Context) {
	gid := ginx.UrlParamInt64(c, "gid")
	lst, err := ltwmodels.DutyRosterWatchkeeperGets(gid, time.Now().Unix())
	ginx.Dangerous(err)

	ginx.NewRender(c).Data(lst, nil)
}
func DutyRosterGets(c *gin.Context) {
	// 接口：根据周数、用户组id获取排班表
	gid := ginx.UrlParamInt64(c, "gid")
	year := ginx.UrlParamInt64(c, "year")
	week := ginx.UrlParamInt64(c, "week")
	var lst []weekDutyRoster

	// 获取用户id和姓名的对应关系
	userMap, err := GetUserMapByGid(gid)
	ginx.Dangerous(err)

	// 获取所选周的时间表
	dt, err := GetWeekIntervalDate(year, week)
	if err != nil {
		ginx.Dangerous(err)
	}

	// 获取所选组的排班配置记录
	dcs, err := ltwmodels.DutyConfGetByGid(gid)
	ginx.Dangerous(err)

	// 返回结果与排班配置记录id的映射关系，方便后面通过排班配置id找到对应的排班记录
	indexMap := make(map[int64]int)

	// 初始化排班记录结果时间表
	for i, dc := range dcs {
		var r weekDutyRoster
		r.DutyConfId = dc.Id
		r.DutyConfName = dc.Name
		r.DutyIntervalTime = dc.StartAt + "~" + dc.EndAt
		r.Monday.DutyDate = dt[0].Date
		r.Tuesday.DutyDate = dt[1].Date
		r.Wednesday.DutyDate = dt[2].Date
		r.Thursday.DutyDate = dt[3].Date
		r.Friday.DutyDate = dt[4].Date
		r.Saturday.DutyDate = dt[5].Date
		r.Sunday.DutyDate = dt[6].Date
		lst = append(lst, r)
		indexMap[dc.Id] = i
	}

	// 获取所选组、所选周所有的值班人员列表
	drs, err := ltwmodels.DutyRosterRecordGets(gid, dt[0].Date, dt[6].Date)
	ginx.Dangerous(err)

	for _, dr := range drs {
		i := indexMap[dr.DutyConfId]

		t, err := time.Parse("20060102", strconv.FormatInt(dr.DutyDate, 10))
		if err != nil {
			ginx.Dangerous(err)
		}
		switch t.Weekday() {
		case time.Monday:
			lst[i].Monday.DutyDate = dr.DutyDate
			lst[i].Monday.Persons = append(
				lst[i].Monday.Persons,
				person{
					UserId:   dr.UserId,
					Username: userMap[dr.UserId],
				})
			break
		case time.Tuesday:
			lst[i].Tuesday.DutyDate = dr.DutyDate
			lst[i].Tuesday.Persons = append(
				lst[i].Tuesday.Persons,
				person{
					UserId:   dr.UserId,
					Username: userMap[dr.UserId],
				})
			break
		case time.Wednesday:
			lst[i].Wednesday.DutyDate = dr.DutyDate
			lst[i].Wednesday.Persons = append(
				lst[i].Wednesday.Persons,
				person{
					UserId:   dr.UserId,
					Username: userMap[dr.UserId],
				})
			break
		case time.Thursday:
			lst[i].Thursday.DutyDate = dr.DutyDate
			lst[i].Thursday.Persons = append(
				lst[i].Thursday.Persons,
				person{
					UserId:   dr.UserId,
					Username: userMap[dr.UserId],
				})
			break
		case time.Friday:
			lst[i].Friday.DutyDate = dr.DutyDate
			lst[i].Friday.Persons = append(
				lst[i].Friday.Persons,
				person{
					UserId:   dr.UserId,
					Username: userMap[dr.UserId],
				})
			break
		case time.Saturday:
			lst[i].Saturday.DutyDate = dr.DutyDate
			lst[i].Saturday.Persons = append(
				lst[i].Saturday.Persons,
				person{
					UserId:   dr.UserId,
					Username: userMap[dr.UserId],
				})
			break
		case time.Sunday:
			lst[i].Sunday.DutyDate = dr.DutyDate
			lst[i].Sunday.Persons = append(
				lst[i].Sunday.Persons,
				person{
					UserId:   dr.UserId,
					Username: userMap[dr.UserId],
				})
			break
		}
	}

	ginx.NewRender(c).Data(lst, nil)
}

func DutyRosterPost(c *gin.Context) {
	// 接口：新增排班记录
	var f operatorForm
	ginx.BindJSON(c, &f)
	gid := ginx.UrlParamInt64(c, "gid")

	// 1.删除历史值班记录
	ltwmodels.DelByDutyAndDate(f.DutyConfId, f.DutyDate)

	// 2.新增最新值班记录
	for _, userId := range f.UserIds {
		dc := ltwmodels.DutyRoster{
			GroupId:    gid,
			DutyConfId: f.DutyConfId,
			UserId:     userId,
			DutyDate:   f.DutyDate,
		}
		dc.Add()
	}

	ginx.NewRender(c).Message(nil)
}

func DutyRosterCopy(c *gin.Context) {
	// 接口：复制排班记录

	//gid := ginx.UrlParamInt64(c, "gid")
	//sy := ginx.QueryInt64(c, "source_year")
	//sw := ginx.QueryInt64(c, "source_week")
	//ty := ginx.QueryInt64(c, "target_week")
	//tw := ginx.QueryInt64(c, "target_week")
	//
	//sd, err := GetWeekIntervalDate(sy, sw)
	//lst, err := ltwmodels.DutyRosterRecordGets(gid, sd[0].Date, ew)
	ginx.NewRender(c).Message("dc.Add()")
}
