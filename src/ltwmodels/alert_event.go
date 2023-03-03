package ltwmodels

//
//import (
//	"bytes"
//	"fmt"
//	"html/template"
//	"strconv"
//	"strings"
//
//	"github.com/didi/nightingale/v5/src/pkg/tplx"
//)
//
//type AlertEvent struct {
//	Id                 int64             `json:"id" gorm:"primaryKey"`
//	Cate               string            `json:"cate"`
//	Cluster            string            `json:"cluster"`
//	GroupId            int64             `json:"group_id"`   // busi group id
//	GroupName          string            `json:"group_name"` // busi group name
//	Hash               string            `json:"hash"`       // rule_id + vector_key
//	RuleId             int64             `json:"rule_id"`
//	RuleName           string            `json:"rule_name"`
//	RuleNote           string            `json:"rule_note"`
//	RuleProd           string            `json:"rule_prod"`
//	RuleAlgo           string            `json:"rule_algo"`
//	Severity           int               `json:"severity"`
//	PromForDuration    int               `json:"prom_for_duration"`
//	PromQl             string            `json:"prom_ql"`
//	PromEvalInterval   int               `json:"prom_eval_interval"`
//	Callbacks          string            `json:"-"`                  // for db
//	CallbacksJSON      []string          `json:"callbacks" gorm:"-"` // for fe
//	RunbookUrl         string            `json:"runbook_url"`
//	NotifyRecovered    int               `json:"notify_recovered"`
//	NotifyChannels     string            `json:"-"`                          // for db
//	NotifyChannelsJSON []string          `json:"notify_channels" gorm:"-"`   // for fe
//	NotifyGroups       string            `json:"-"`                          // for db
//	NotifyGroupsJSON   []string          `json:"notify_groups" gorm:"-"`     // for fe
//	NotifyGroupsObj    []*UserGroup      `json:"notify_groups_obj" gorm:"-"` // for fe
//	TargetIdent        string            `json:"target_ident"`
//	TargetNote         string            `json:"target_note"`
//	TriggerTime        int64             `json:"trigger_time"`
//	TriggerValue       string            `json:"trigger_value"`
//	Tags               string            `json:"-"`                         // for db
//	TagsJSON           []string          `json:"tags" gorm:"-"`             // for fe
//	TagsMap            map[string]string `json:"-" gorm:"-"`                // for internal usage
//	IsRecovered        bool              `json:"is_recovered" gorm:"-"`     // for notify.py
//	NotifyUsersObj     []*User           `json:"notify_users_obj" gorm:"-"` // for notify.py
//	LastEvalTime       int64             `json:"last_eval_time" gorm:"-"`   // for notify.py 上次计算的时间
//	LastSentTime       int64             `json:"last_sent_time" gorm:"-"`   // 上次发送时间
//	NotifyNumber       int               `json:"notify__number"`            // notify: rent number
//	FirstTriggerTime   int64             `json:"first_trigger_time"`        // 连续告警的首次告警时间
//}
//
//func (e *AlertEvent) TableName() string {
//	return "ltw_alert_event"
//}
//
//func (e *AlertEvent) Add() error {
//	return Insert(e)
//}
//
//type AggrRule struct {
//	Type  string
//	Value string
//}
//
//func (e *AlertEvent) ParseRule(field string) error {
//	f := e.GetField(field)
//	f = strings.TrimSpace(f)
//
//	if f == "" {
//		return nil
//	}
//
//	var defs = []string{
//		"{{$labels := .TagsMap}}",
//		"{{$value := .TriggerValue}}",
//	}
//
//	text := strings.Join(append(defs, f), "")
//	t, err := template.New(fmt.Sprint(e.RuleId)).Funcs(tplx.TemplateFuncMap).Parse(text)
//	if err != nil {
//		return err
//	}
//
//	var body bytes.Buffer
//	err = t.Execute(&body, e)
//	if err != nil {
//		return err
//	}
//
//	if field == "rule_name" {
//		e.RuleName = body.String()
//	}
//
//	if field == "rule_note" {
//		e.RuleNote = body.String()
//	}
//	return nil
//}
//
//func (e *AlertEvent) GenCardTitle(rules []*AggrRule) string {
//	arr := make([]string, len(rules))
//	for i := 0; i < len(rules); i++ {
//		rule := rules[i]
//
//		if rule.Type == "field" {
//			arr[i] = e.GetField(rule.Value)
//		}
//
//		if rule.Type == "tagkey" {
//			arr[i] = e.GetTagValue(rule.Value)
//		}
//
//		if len(arr[i]) == 0 {
//			arr[i] = "Null"
//		}
//	}
//	return strings.Join(arr, "::")
//}
//
//func (e *AlertEvent) GetTagValue(tagkey string) string {
//	for _, tag := range e.TagsJSON {
//		i := strings.Index(tag, tagkey+"=")
//		if i >= 0 {
//			return tag[len(tagkey+"="):]
//		}
//	}
//	return ""
//}
//
//func (e *AlertEvent) GetField(field string) string {
//	switch field {
//	case "cluster":
//		return e.Cluster
//	case "group_id":
//		return fmt.Sprint(e.GroupId)
//	case "group_name":
//		return e.GroupName
//	case "rule_id":
//		return fmt.Sprint(e.RuleId)
//	case "rule_name":
//		return e.RuleName
//	case "rule_note":
//		return e.RuleNote
//	case "severity":
//		return fmt.Sprint(e.Severity)
//	case "runbook_url":
//		return e.RunbookUrl
//	case "target_ident":
//		return e.TargetIdent
//	case "target_note":
//		return e.TargetNote
//	default:
//		return ""
//	}
//}
//
//func (e *AlertEvent) DB2FE() {
//	e.NotifyChannelsJSON = strings.Fields(e.NotifyChannels)
//	e.NotifyGroupsJSON = strings.Fields(e.NotifyGroups)
//	e.CallbacksJSON = strings.Fields(e.Callbacks)
//	e.TagsJSON = strings.Split(e.Tags, ",,")
//}
//
//func (e *AlertEvent) DB2Mem() {
//	e.IsRecovered = false
//	e.NotifyGroupsJSON = strings.Fields(e.NotifyGroups)
//	e.CallbacksJSON = strings.Fields(e.Callbacks)
//	e.NotifyChannelsJSON = strings.Fields(e.NotifyChannels)
//	e.TagsJSON = strings.Split(e.Tags, ",,")
//	e.TagsMap = make(map[string]string)
//	for i := 0; i < len(e.TagsJSON); i++ {
//		pair := strings.TrimSpace(e.TagsJSON[i])
//		if pair == "" {
//			continue
//		}
//
//		arr := strings.Split(pair, "=")
//		if len(arr) != 2 {
//			continue
//		}
//
//		e.TagsMap[arr[0]] = arr[1]
//	}
//}
//
//// for webui
//func (e *AlertEvent) FillNotifyGroups(cache map[int64]*UserGroup) error {
//	// some user-group already deleted ?
//	count := len(e.NotifyGroupsJSON)
//	if count == 0 {
//		e.NotifyGroupsObj = []*UserGroup{}
//		return nil
//	}
//
//	for i := range e.NotifyGroupsJSON {
//		id, err := strconv.ParseInt(e.NotifyGroupsJSON[i], 10, 64)
//		if err != nil {
//			continue
//		}
//
//		ug, has := cache[id]
//		if has {
//			e.NotifyGroupsObj = append(e.NotifyGroupsObj, ug)
//			continue
//		}
//
//		ug, err = UserGroupGetById(id)
//		if err != nil {
//			return err
//		}
//
//		if ug != nil {
//			e.NotifyGroupsObj = append(e.NotifyGroupsObj, ug)
//			cache[id] = ug
//		}
//	}
//
//	return nil
//}
//
//func AlertEventTotal(prod string, bgid, stime, etime int64, severity int, clusters, cates []string, query string) (int64, error) {
//	session := DB().Model(&AlertEvent{}).Where("trigger_time between ? and ? and rule_prod = ?", stime, etime, prod)
//
//	if bgid > 0 {
//		session = session.Where("group_id = ?", bgid)
//	}
//
//	if severity >= 0 {
//		session = session.Where("severity = ?", severity)
//	}
//
//	if len(clusters) > 0 {
//		session = session.Where("cluster in ?", clusters)
//	}
//
//	if len(cates) > 0 {
//		session = session.Where("cate in ?", cates)
//	}
//
//	if query != "" {
//		arr := strings.Fields(query)
//		for i := 0; i < len(arr); i++ {
//			qarg := "%" + arr[i] + "%"
//			session = session.Where("rule_name like ? or tags like ?", qarg, qarg)
//		}
//	}
//
//	return Count(session)
//}
//
//func AlertEventGets(prod string, bgid, stime, etime int64, severity int, clusters, cates []string, query string, limit, offset int) ([]AlertEvent, error) {
//	session := DB().Where("trigger_time between ? and ? and rule_prod = ?", stime, etime, prod)
//
//	if bgid > 0 {
//		session = session.Where("group_id = ?", bgid)
//	}
//
//	if severity >= 0 {
//		session = session.Where("severity = ?", severity)
//	}
//
//	if len(clusters) > 0 {
//		session = session.Where("cluster in ?", clusters)
//	}
//
//	if len(cates) > 0 {
//		session = session.Where("cate in ?", cates)
//	}
//
//	if query != "" {
//		arr := strings.Fields(query)
//		for i := 0; i < len(arr); i++ {
//			qarg := "%" + arr[i] + "%"
//			session = session.Where("rule_name like ? or tags like ?", qarg, qarg)
//		}
//	}
//
//	var lst []AlertEvent
//	err := session.Order("id desc").Limit(limit).Offset(offset).Find(&lst).Error
//
//	if err == nil {
//		for i := 0; i < len(lst); i++ {
//			lst[i].DB2FE()
//		}
//	}
//
//	return lst, err
//}
//
//func AlertEventDel(ids []int64) error {
//	if len(ids) == 0 {
//		return nil
//	}
//
//	return DB().Where("id in ?", ids).Delete(&AlertEvent{}).Error
//}
//
//func AlertEventDelByHash(hash string) error {
//	return DB().Where("hash = ?", hash).Delete(&AlertEvent{}).Error
//}
//
//func AlertEventExists(where string, args ...interface{}) (bool, error) {
//	return Exists(DB().Model(&AlertEvent{}).Where(where, args...))
//}
//
//func AlertEventGet(where string, args ...interface{}) (*AlertEvent, error) {
//	var lst []*AlertEvent
//	err := DB().Where(where, args...).Find(&lst).Error
//	if err != nil {
//		return nil, err
//	}
//
//	if len(lst) == 0 {
//		return nil, nil
//	}
//
//	lst[0].DB2FE()
//	lst[0].FillNotifyGroups(make(map[int64]*UserGroup))
//
//	return lst[0], nil
//}
//
//func AlertEventGetById(id int64) (*AlertEvent, error) {
//	return AlertEventGet("id=?", id)
//}
//
//type AlertNumber struct {
//	GroupId    int64
//	GroupCount int64
//}
//
//// for busi_group list page
//func AlertNumbers(bgids []int64) (map[int64]int64, error) {
//	ret := make(map[int64]int64)
//	if len(bgids) == 0 {
//		return ret, nil
//	}
//
//	var arr []AlertNumber
//	err := DB().Model(&AlertEvent{}).Select("group_id", "count(*) as group_count").Where("group_id in ?", bgids).Group("group_id").Find(&arr).Error
//	if err != nil {
//		return nil, err
//	}
//
//	for i := 0; i < len(arr); i++ {
//		ret[arr[i].GroupId] = arr[i].GroupCount
//	}
//
//	return ret, nil
//}
//
//func AlertEventGetAll(cluster string) ([]*AlertEvent, error) {
//	session := DB().Model(&AlertEvent{})
//
//	if cluster != "" {
//		session = session.Where("cluster = ?", cluster)
//	}
//
//	var lst []*AlertEvent
//	err := session.Find(&lst).Error
//	return lst, err
//}
//
//func AlertEventGetByIds(ids []int64) ([]*AlertEvent, error) {
//	var lst []*AlertEvent
//
//	if len(ids) == 0 {
//		return lst, nil
//	}
//
//	err := DB().Where("id in ?", ids).Order("id desc").Find(&lst).Error
//	if err == nil {
//		for i := 0; i < len(lst); i++ {
//			lst[i].DB2FE()
//		}
//	}
//
//	return lst, err
//}
//
//func AlertEventGetByRule(ruleId int64) ([]*AlertEvent, error) {
//	var lst []*AlertEvent
//	err := DB().Where("rule_id=?", ruleId).Find(&lst).Error
//	return lst, err
//}
//
//func AlertEventGetMap(cluster string) (map[int64]map[string]struct{}, error) {
//	session := DB().Model(&AlertEvent{})
//	if cluster != "" {
//		session = session.Where("cluster = ?", cluster)
//	}
//
//	var lst []*AlertEvent
//	err := session.Select("rule_id", "hash").Find(&lst).Error
//	if err != nil {
//		return nil, err
//	}
//
//	ret := make(map[int64]map[string]struct{})
//	for i := 0; i < len(lst); i++ {
//		rid := lst[i].RuleId
//		hash := lst[i].Hash
//		if _, has := ret[rid]; has {
//			ret[rid][hash] = struct{}{}
//		} else {
//			ret[rid] = make(map[string]struct{})
//			ret[rid][hash] = struct{}{}
//		}
//	}
//
//	return ret, nil
//}
