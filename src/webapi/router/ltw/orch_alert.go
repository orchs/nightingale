package ltw

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/didi/nightingale/v5/src/models"
	"github.com/didi/nightingale/v5/src/pkg/ltw"
	"github.com/didi/nightingale/v5/src/storage"
	"github.com/didi/nightingale/v5/src/webapi/config"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/toolkits/pkg/ginx"
	"github.com/toolkits/pkg/logger"
	"net/http"
	"strconv"
	"time"
)

const LastUpdatePrefix = "ORCH_UPDATE_TIME_"

func initParams(gid int64) (*config.LtwORCHEnvInfo, error) {
	for _, c := range config.C.Ltw.ORCHEnvInfos {
		if c.GroupId == gid {
			return &c, nil
		}
	}
	return nil, errors.New(fmt.Sprintf("目标业务组没有配置，gid：%v", gid))
}

func SyncOAlert(c *gin.Context) {
	// 通过categraf周期性调用

	// 1.init params
	gidStr := ginx.QueryStr(c, "gid")
	gid, _ := strconv.ParseInt(gidStr, 10, 64)
	cf, err := initParams(gid)
	if err != nil {
		ginx.Bomb(http.StatusInternalServerError, "参数错误: %s", err)
	}

	// 2.登录，获取token
	token, err := oauthToken(cf.Domain, config.C.Ltw.ORCHGrantType, cf.ClientId, cf.ClientSecret, cf.GroupName)
	if err != nil {
		logger.Errorf("获取 %v token失败！ 错误信息：%v, 响应数据：%s", cf.GroupName, err)
		ginx.Bomb(http.StatusInternalServerError, "登录错误: %s", err)
	}

	// 3.拉取最新时间段内告警数据
	now := time.Now()
	bd, _ := time.ParseDuration("-8h00m30s")
	before := now.Add(bd).Format("2006-01-02T03:04:05Z")

	ctx := context.Background()
	lastUpdateTIme := LastUpdatePrefix + gidStr
	after, err := storage.Redis.Get(ctx, lastUpdateTIme).Result()
	if err != nil {
		d, _ := time.ParseDuration("-9h00m30s")
		after = now.Add(d).Format("2006-01-02T03:04:05Z")
	}

	logger.Debugf("开始同步%v数据,同步范围 %v ~ %v", cf.GroupName, after, before)
	alerts, err := getOAlerts(after, before, cf.Domain, token)
	if err != nil {
		ginx.Bomb(http.StatusInternalServerError, "拉取%v数据出错:%s", cf.GroupName, err)
	}
	curAlerts, err := getLocalCurAlerts(gid, cf.GroupName, cf.Domain, token)
	if err != nil {
		ginx.Bomb(http.StatusInternalServerError, "同步活跃告警信息%v出错:%v", cf.GroupName, err)
	}

	alerts = append(alerts, curAlerts...)
	// 4.分析、保存告警数据
	handleAlerts(gid, cf.GroupName, alerts)

	storage.Redis.Set(ctx, lastUpdateTIme, before, time.Duration(time.Hour)).Err()
	logger.Debugf("%v同步完成：同步范围 %v ~ %v", cf.GroupName, after, before)
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
		} else {
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

func getOAlerts(after, before, domain, token string) ([]ORCHAlertContent, error) {
	api := domain + config.C.Ltw.ORCHAlertsApi
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
		return nil, err
	}

	return &data, nil
}

func oauthToken(domain, gt, cid, cs, gName string) (string, error) {
	//ctx := context.Background()
	//tokenName := "o_token_" + cid
	//
	//token, err := storage.Redis.Get(ctx, tokenName).Result()
	//if err == nil {
	//	return token, nil
	//}

	tokenApi := domain + "/oauth/token"
	readerStr := fmt.Sprintf("grant_type=%v&client_id=%v&client_secret=%v", gt, cid, cs)

	body, err := ltw.HttpPost(tokenApi, readerStr)
	if err != nil {
		logger.Errorf("登录%v失败！%v, 错误信息：%v", gName, tokenApi, err)
		return "", err
	}

	logger.Errorf("登录%v, 请求：%v, 响应信息：%s", gName, tokenApi, body)

	var data ORCHAuthResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return "", errors.Wrapf(err, "获取 %v token失败！%v, 错误信息：%v, 响应数据：%s", gName, tokenApi, err, body)
	}

	token := data.TokenType + " " + data.AccessToken
	//expiresIn := time.Duration(time.Second * time.Duration(data.ExpiresIn))
	//storage.Redis.Set(ctx, tokenName, token, expiresIn).Err()

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
	api := domain + config.C.Ltw.ORCHAlertApi
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
