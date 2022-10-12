package ltw

import (
	"fmt"
	"github.com/didi/nightingale/v5/src/ltwmodels"
	"github.com/gin-gonic/gin"
	"github.com/toolkits/pkg/ginx"
	"strconv"
	"time"
)

type WeekRangeDate struct {
	Date    int64  `json:"date"`
	Week    int64  `json:"week"`
	Display string `json:"display"`
}

var WeekMap = map[int]string{
	1: "周一",
	2: "周二",
	3: "周三",
	4: "周四",
	5: "周五",
	6: "周六",
	7: "周日",
}

func DutyConfPut(c *gin.Context) {
	var f ltwmodels.DutyConf
	ginx.BindJSON(c, &f)

	cid := ginx.UrlParamInt64(c, "cid")
	dc, err := ltwmodels.DutyConfGetById(cid)
	ginx.Dangerous(err)

	ginx.NewRender(c).Message(dc.Update(f))
}

func DutyConfDel(c *gin.Context) {
	cid := ginx.UrlParamInt64(c, "cid")
	dc, err := ltwmodels.DutyConfGetById(cid)
	ginx.Dangerous(err)

	ginx.NewRender(c).Message(dc.Delete())
}

func DutyConfGets(c *gin.Context) {
	// 接口：获取班次列表
	gid := ginx.UrlParamInt64(c, "gid")

	lst, err := ltwmodels.DutyConfGetByGid(gid)
	ginx.Dangerous(err)

	ginx.NewRender(c).Data(lst, nil)
}

func DutyConfAdd(c *gin.Context) {
	// 接口：新增班次
	var dc ltwmodels.DutyConf
	ginx.BindJSON(c, &dc)

	gid := ginx.UrlParamInt64(c, "gid")
	dc.GroupId = gid
	err := dc.Add()
	ginx.Dangerous(err)

	ginx.NewRender(c).Message("")
}

func GetWeekIntervalDate(year, week int64) ([]WeekRangeDate, error) {
	// 接口：根据年和周次获取日期
	var ts []WeekRangeDate
	t := WeekStart(year, week)
	for i := 0; i <= 6; i++ {
		tt := t.AddDate(0, 0, i)
		s := tt.Format("20060102")
		d, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return nil, err
		}
		md := fmt.Sprintf("%v/%v", WeekMap[i+1], tt.Format("1.2"))

		ts = append(ts, WeekRangeDate{d, int64(i) + 1, md})
	}
	return ts, nil
}

func WeekStart(year, week int64) time.Time {
	// Start from the middle of the year:
	t := time.Date(int(year), 7, 1, 0, 0, 0, 0, time.UTC)

	// Roll back to Monday:
	if wd := t.Weekday(); wd == time.Sunday {
		t = t.AddDate(0, 0, -6)
	} else {
		t = t.AddDate(0, 0, -int(wd)+1)
	}

	// Difference in weeks:
	_, w := t.ISOWeek()
	t = t.AddDate(0, 0, (int(week)-w)*7)

	return t
}

func GetWeekDays(c *gin.Context) {
	// 接口：根据年、周获取日期表头
	year := ginx.UrlParamInt64(c, "year")
	week := ginx.UrlParamInt64(c, "week")

	t, err := GetWeekIntervalDate(year, week)
	ginx.Dangerous(err)
	ginx.NewRender(c).Data(t, nil)
}
