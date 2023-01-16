package ltw

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/didi/nightingale/v5/src/models"
	"github.com/didi/nightingale/v5/src/pkg/ltw"
	"github.com/didi/nightingale/v5/src/storage"
	"github.com/gin-gonic/gin"
	"github.com/toolkits/pkg/ginx"
	"github.com/toolkits/pkg/logger"
	"strconv"
	"time"
)

const LastUpdatePrefix = "ORCH_UPDATE_TIME_"
const ORCHAlertsApi = "/openapi/v2/alerts"
const ORCHAlertApi = "/openapi/v2/alert/"

var groupMap = map[int64]string{
	1001: "orch_国内_告警",
	1002: "orch_微软_告警",
	1003: "orch_平安办_告警",
	1004: "orch_香港_告警",
	1005: "orch_国电投_告警",
	1006: "orch_中交_告警",
	1007: "orch_中国电子_告警",
	1008: "orch_中能建_告警",
	1009: "orch_中电科_告警",
	1010: "orch_万达_告警",
	1011: "orch_中烟_告警",
	1012: "orch_有研_告警",
	1013: "orch_有矿_告警",
	1014: "orch_中交海外组网_告警",
}

func SyncOAlert(c *gin.Context) {
	// 通过categraf周期性调用
	gt := ginx.QueryStr(c, "gt")
	domain := ginx.QueryStr(c, "domain")
	cid := ginx.QueryStr(c, "cid")
	cs := ginx.QueryStr(c, "cs")
	gidStr := ginx.QueryStr(c, "gid")
	env := ginx.QueryStr(c, "env")

	// 1.检查传参
	if domain == "" || gt == "" || cid == "" || cs == "" || gidStr == "" || env == "" {
		ginx.Bomb(400, "传参错误！")
	}
	gid, err := strconv.ParseInt(gidStr, 10, 64)
	if err != nil {
		ginx.Bomb(500, "gid错误！")
	}
	if _, ok := groupMap[gid]; !ok {
		ginx.Bomb(500, "gid不支持！")
	}
	gName := groupMap[gid]

	// 2.登录，获取token
	token, err := oauthToken(domain, gt, cid, cs, env)
	if err != nil {
		ginx.Bomb(500, "登录错误！")
	}

	// 3.拉取最新时间段内告警数据

	now := time.Now()
	bd, _ := time.ParseDuration("-8h00m30s")
	before := now.Add(bd).Format("2006-01-02T03:04:05Z")

	ctx := context.Background()
	lastUpdateTIme := LastUpdatePrefix + domain
	after, err := storage.Redis.Get(ctx, lastUpdateTIme).Result()
	if err != nil {
		d, _ := time.ParseDuration("-9h00m30s")
		after = now.Add(d).Format("2006-01-02T03:04:05Z")
	}

	logger.Debugf("开始同步%v数据,同步范围 %v ~ %v", env, after, before)
	alerts, err := getOAlerts(after, before, domain, env, token)
	if err != nil {
		ginx.Bomb(500, fmt.Sprintf("拉取%v数据出错:%v", env, err))
	}
	curAlerts, err := getLocalCurAlerts(gid, gName, domain, token)
	if err != nil {
		ginx.Bomb(500, fmt.Sprintf("同步活跃告警信息%v出错:%v", env, err))
	}

	alerts = append(alerts, curAlerts...)
	// 4.分析、保存告警数据
	handleAlerts(gid, gName, alerts)

	storage.Redis.Set(ctx, lastUpdateTIme, before, time.Duration(time.Hour)).Err()
	ginx.NewRender(c).Data("", nil)
}

func handleAlerts(gid int64, gName string, alerts []ORCHAlertContent) {

	for _, v := range alerts {
		tags := fmt.Sprintf("target_name=%v,,target_type=%v,,rule_name=%v,,details=%v", v.TargetName, v.TargetType, v.Type, v.Details)
		event := models.AlertCurEvent{
			Cluster:     "Default",
			GroupId:     gid,
			GroupName:   gName,
			Hash:        v.Id,
			RuleId:      0,
			RuleName:    v.Type,
			RuleNote:    v.Details,
			Severity:    SeverityMap[v.Level],
			TargetIdent: v.TargetName,
			IsRecovered: false,
			Tags:        tags,
		}
		if len(v.Events) == 1 {
			firstTime, _ := time.ParseInLocation("2006-01-02 15:04:05", v.Events[0].CreatedAt, time.Local)
			event.FirstTriggerTime = firstTime.Unix()
			event.TriggerTime = firstTime.Unix()
		} else if v.Events[0].Event == "closed" {
			firstTime, _ := time.ParseInLocation("2006-01-02 15:04:05", v.Events[len(v.Events)-1].CreatedAt, time.Local)
			curTime, _ := time.ParseInLocation("2006-01-02 15:04:05", v.Events[0].CreatedAt, time.Local)
			event.FirstTriggerTime = firstTime.Unix()
			event.TriggerTime = firstTime.Unix()
			event.LastEvalTime = curTime.Unix()
			event.IsRecovered = true
		}
		persist(&event)
	}
}

func getOAlerts(after, before, domain, env, token string) ([]ORCHAlertContent, error) {
	api := fmt.Sprintf("https://%v%v", domain, ORCHAlertsApi)
	res, err := reqOAlerts(api, token, after, before, "0")
	if err != nil {
		return nil, err
	}

	var alerts = res.Content

	if res.TotalPages == 1 {
		return alerts, nil
	}

	for i := 1; i < int(res.TotalPages); i++ {
		res, err := reqOAlerts(api, token, after, before, string(i))
		if err != nil {
			return nil, err
		}
		alerts = append(alerts, res.Content...)
	}

	return alerts, nil
}

func reqOAlerts(api, token, after, before, page string) (*ORCHAlertsResponse, error) {

	res, err := ltw.HttpGet(
		api,
		map[string]string{
			"after":  after,
			"before": before,
			"size":   "100",
			"page":   page,
		},
		map[string]string{
			"Authorization": token,
		},
	)
	if err != nil {
		return nil, err
	}

	var data ORCHAlertsResponse
	if err := json.Unmarshal(res, &data); err != nil {
		logger.Errorf("数据解析失败！data: %v, 错误信息：%v", string(res), err)
		return nil, err
	}

	return &data, nil
}
func oauthToken(domain, gt, cid, cs, env string) (string, error) {
	ctx := context.Background()
	tokenName := "o_token_" + cid

	token, err := storage.Redis.Get(ctx, tokenName).Result()
	if err == nil {
		return token, nil
	}

	tokenApi := fmt.Sprintf("https://%v/oauth/token", domain)
	readerStr := fmt.Sprintf("grant_type=%v&client_id=%v&client_secret=%v", gt, cid, cs)

	body, err := ltw.HttpPost(tokenApi, readerStr)
	if err != nil {
		logger.Errorf("登录%v失败！%v, 错误信息：%v", env, tokenApi, err)
		return "", err
	}

	var data ORCHAuthResponse
	if err := json.Unmarshal(body, &data); err != nil {
		logger.Errorf("获取 %v token失败！%v, 错误信息：%v", env, tokenApi, err)
		return "", err
	}

	token = data.TokenType + " " + data.AccessToken
	expiresIn := time.Duration(time.Second * time.Duration(data.ExpiresIn))
	storage.Redis.Set(ctx, tokenName, token, expiresIn).Err()

	return token, nil
}

func getLocalCurAlerts(gid int64, gName, domain, token string) ([]ORCHAlertContent, error) {
	// 1.根据group_id,rule_id查询活跃告警
	now := time.Now()
	var sTime int64 = now.AddDate(0, 0, -1).Unix()
	eTime := now.Unix()
	list, err := models.AlertCurEventGets("", gid, sTime, eTime, -1, nil, nil, "", 1000, 0)
	if err != nil {
		return nil, err
	}

	// 2.遍历搜集到的活跃告警，给对应的orch发送请求，携带status字段，只查找已经恢复的告警
	api := fmt.Sprintf("https://%v%v", domain, ORCHAlertApi)
	var hisAlerts []ORCHAlertContent
	for _, v := range list {
		a, err := ltw.HttpGet(
			api+v.Hash,
			nil,
			map[string]string{
				"Authorization": token,
			},
		)
		if err != nil {
			logger.Errorf("%v 获取 %v 告警信息失败！错误信息：%v", gName, v.Hash, err)
			continue
		}

		var data ORCHAlertContent

		if err := json.Unmarshal(a, &data); err != nil {
			logger.Errorf("%v 数据解析失败！data: %v 错误信息：%v", gName, string(a), err)
			continue
		}

		if data.Status == "closed" {
			hisAlerts = append(hisAlerts, data)
		}
	}

	return hisAlerts, nil
}
