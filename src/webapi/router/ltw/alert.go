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
	"github.com/toolkits/pkg/ginx"
	"github.com/toolkits/pkg/logger"
	"time"
)

const LastUpdateTimeTag = "orch_last_update_time"

type ORCHAlertEvent struct {
	Id        string `json:"id"`
	CreatedAt string `json:"created_at"`
	Event     string `json:"event"`
	Operator  string `json:"operator"`
	AlertId   string `json:"alert_id"`
}

type ORCHAlertContent struct {
	Id         string           `json:"id"`
	TargetName string           `json:"target_name"`
	Level      string           `json:"level"`
	Status     string           `json:"status"`
	Details    string           `json:"details"`
	Type       string           `json:"type"`
	TargetType string           `json:"target_type"`
	TargetKind string           `json:"target_kind"`
	TargetId   string           `json:"target_id"`
	Events     []ORCHAlertEvent `json:"events"`
}

type ORCHAlertsResponse struct {
	TotalElements int64              `json:"total_elements"`
	TotalPages    int64              `json:"total_pages"`
	Number        int64              `json:"number"`
	Size          int64              `json:"size"`
	Content       []ORCHAlertContent `json:"content"`
}

type ORCHAuthResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
}

var SeverityMap = map[string]int{
	"critical": 1,
	"major":    2,
	"trivial":  3,
}

func SyncOrchAlert(c *gin.Context) {
	// 通过categraf周期性调用
	d := ginx.UrlParamStr(c, "d")
	cid := ginx.UrlParamStr(c, "cid")
	cs := ginx.UrlParamStr(c, "cs")
	gid := ginx.UrlParamInt64(c, "gid")
	env := ginx.UrlParamStr(c, "env")
	if d == "" || cid == "" || cs == "" || gid == 0 || env == "" {
		ginx.Bomb(400, "传参错误！")
	}

	token, err := getOrchToken(d, cid, cs, env)
	if err != nil {
		ginx.Bomb(500, "登录错误！")
	}
	getOrchAlert(d, token)
	ginx.NewRender(c).Data("", nil)
}

func getOrchAlert(d, token string) {

}
func getOrchToken(d, cid, cs, env string) (string, error) {
	ctx := context.Background()
	tokenName := "o_token_" + env

	token, err := storage.Redis.Get(ctx, tokenName).Result()
	if err == nil {
		return token, nil
	}

	readerStr := fmt.Sprintf("grant_type=%v&client_id=%v&client_secret=%v", config.C.Ltw.ORCHGrantType, cid, cs)

	tokenApi := d + config.C.Ltw.ORCHTokenApi
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

func ORCHAlertUpdateTask() {
	// 执行同步任务
	now := time.Now()
	bd, _ := time.ParseDuration("-8h00m00s")
	before := now.Add(bd).Format("2006-01-02T03:04:05Z")

	ctx := context.Background()
	after, err := storage.Redis.Get(ctx, LastUpdateTimeTag).Result()
	if err != nil {
		d, _ := time.ParseDuration("-9h00m15s")
		after = now.Add(d).Format("2006-01-02T03:04:05Z")
	}
	storage.Redis.Set(ctx, LastUpdateTimeTag, before, time.Duration(time.Hour)).Err()

	logger.Debugf("开始新一轮数据同步,同步范围 %v ~ %v", after, before)
	for _, c := range config.C.Ltw.ORCHEnvInfos {
		if !c.Apply {
			continue
		}
		go updateORCHAlert(
			ctx,
			c,
			map[string]string{
				"after":  after,
				"before": before,
				"size":   "100",
			},
		)
	}
}

func checkLocalCurAlerts(token string, conf config.LtwORCHEnvInfo) []ORCHAlertContent {
	// 1.根据group_id,rule_id查询活跃告警
	now := time.Now()
	var sTime int64 = now.AddDate(0, 0, -30).Unix()
	eTime := now.Unix()
	list, err := models.AlertCurEventGets("", conf.GroupId, sTime, eTime, -1, nil, nil, "", 1000, 0)
	if err != nil {
		logger.Errorf("%v查询活跃告警出错：%v", conf.GroupName, err)

	}

	var hisAlerts []ORCHAlertContent
	// 2.遍历搜集到的活跃告警，给对应的orch发送请求，携带status字段，只查找已经恢复的告警
	for _, v := range list {
		a, err := ltw.HttpGet(
			conf.Domain+config.C.Ltw.ORCHAlertApi+v.Hash,
			nil,
			map[string]string{
				"Authorization": token,
			},
		)
		if err != nil {
			logger.Errorf("%v 获取 %v 告警信息失败！错误信息：%v", conf.GroupName, v.Hash, err)
			continue
		}

		var data ORCHAlertContent
		if err := json.Unmarshal(a, &data); err != nil {
			logger.Errorf("%v 数据解析失败！data: %v 错误信息：%v", conf.GroupName, a, err)
			continue
		}

		if data.Events == nil {
			logger.Errorf("%v 未找到对应告警！错误信息：%v", conf.GroupName, v.Hash)
			continue
		}

		if data.Events[0].Event != "firing" {
			hisAlerts = append(hisAlerts, data)
		}
	}
	return hisAlerts
}

func updateORCHAlert(ctx context.Context, conf config.LtwORCHEnvInfo, params map[string]string) {

	logger.Debugf("开始同步 %v 告警数据。。。 。。。", conf.GroupName)

	token, err := getOToken(ctx, conf)
	if err != nil {
		return
	}

	alerts := checkLocalCurAlerts(token, conf)

	lst, err := ltw.HttpGet(
		conf.Domain+config.C.Ltw.ORCHAlertsApi,
		params,
		map[string]string{
			"Authorization": token,
		},
	)
	var data ORCHAlertsResponse
	if err := json.Unmarshal(lst, &data); err != nil {
		logger.Errorf("%v 数据解析失败！data: %v, 错误信息：%v", conf.GroupName, lst, err)
		return
	}

	alerts = append(alerts, data.Content...)
	for _, v := range alerts {
		tags := fmt.Sprintf("target_name=%v,,target_type=%v,,rule_name=%v,,details=%v", v.TargetName, v.TargetType, v.Type, v.Details)
		event := models.AlertCurEvent{
			Cluster:     "Default",
			GroupId:     conf.GroupId,
			GroupName:   conf.GroupName,
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
			firstTime, _ := time.ParseInLocation("2006-01-02 15:04:05", v.Events[1].CreatedAt, time.Local)
			curTime, _ := time.ParseInLocation("2006-01-02 15:04:05", v.Events[0].CreatedAt, time.Local)
			event.FirstTriggerTime = firstTime.Unix()
			event.TriggerTime = firstTime.Unix()
			event.LastEvalTime = curTime.Unix()
			event.IsRecovered = true
		}
		persist(&event)
	}
	logger.Debugf("%v 数据同步完成！", conf.GroupName)
}

func getOToken(ctx context.Context, conf config.LtwORCHEnvInfo) (string, error) {
	tokenName := "o_token_" + string(conf.GroupId)

	token, err := storage.Redis.Get(ctx, tokenName).Result()
	if err == nil {
		return token, nil
	}

	readerStr := fmt.Sprintf(
		"grant_type=%v&client_id=%v&client_secret=%v",
		config.C.Ltw.ORCHGrantType,
		conf.ClientId,
		conf.ClientSecret,
	)

	tokenApi := conf.Domain + config.C.Ltw.ORCHTokenApi
	body, err := ltw.HttpPost(tokenApi, readerStr)
	if err != nil {
		logger.Errorf("登录%v失败！%v, 错误信息：%v", conf.GroupName, tokenApi, err)
		return "", err
	}

	var data ORCHAuthResponse
	if err := json.Unmarshal(body, &data); err != nil {
		logger.Errorf("获取 %v token失败！%v, 错误信息：%v", conf.GroupName, tokenApi, err)
		return "", err
	}

	token = data.TokenType + " " + data.AccessToken
	expiresIn := time.Duration(time.Second * time.Duration(data.ExpiresIn))
	storage.Redis.Set(ctx, tokenName, token, expiresIn).Err()

	return token, nil
}

func persist(event *models.AlertCurEvent) {

	has, err := models.AlertCurEventExists("hash=?", event.Hash)
	if err != nil {
		logger.Errorf("event_persist_check_exists_fail: %v rule_id=%d hash=%s", err, event.RuleId, event.Hash)
		return
	}

	his := event.ToHis()

	// 不管是告警还是恢复，全量告警里都要记录
	if err := his.Add(); err != nil {
		logger.Errorf(
			"event_persist_his_fail: %v rule_id=%d hash=%s tags=%v timestamp=%d value=%s",
			err,
			event.RuleId,
			event.Hash,
			event.TagsJSON,
			event.TriggerTime,
			event.TriggerValue,
		)
	}

	if has {
		// 活跃告警表中有记录，删之
		err = models.AlertCurEventDelByHash(event.Hash)
		if err != nil {
			logger.Errorf("event_del_cur_fail: %v hash=%s", err, event.Hash)
			return
		}

		if !event.IsRecovered {
			// 恢复事件，从活跃告警列表彻底删掉，告警事件，要重新加进来新的event
			// use his id as cur id
			event.Id = his.Id
			if event.Id > 0 {
				if err := event.Add(); err != nil {
					logger.Errorf(
						"event_persist_cur_fail: %v rule_id=%d hash=%s tags=%v timestamp=%d value=%s",
						err,
						event.RuleId,
						event.Hash,
						event.TagsJSON,
						event.TriggerTime,
						event.TriggerValue,
					)
				}
			}
		}

		return
	}

	if event.IsRecovered {
		// alert_cur_event表里没有数据，表示之前没告警，结果现在报了恢复，神奇....理论上不应该出现的
		return
	}

	// use his id as cur id
	event.Id = his.Id
	if event.Id > 0 {
		if err := event.Add(); err != nil {
			logger.Errorf(
				"event_persist_cur_fail: %v rule_id=%d hash=%s tags=%v timestamp=%d value=%s",
				err,
				event.RuleId,
				event.Hash,
				event.TagsJSON,
				event.TriggerTime,
				event.TriggerValue,
			)
		}
	}
}

func PullORCHAlertStart() {
	for {
		time.Sleep(time.Second * time.Duration(config.C.Ltw.ORCHPullInterval))
		ORCHAlertUpdateTask()
	}
}
