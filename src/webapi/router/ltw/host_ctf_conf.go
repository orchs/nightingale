package ltw

import (
	"errors"
	"github.com/didi/nightingale/v5/src/ltwmodels"
	"github.com/didi/nightingale/v5/src/models"
	"github.com/didi/nightingale/v5/src/pkg/ltw"
	"github.com/didi/nightingale/v5/src/pkg/ltw/ctf"
	"github.com/gammazero/workerpool"
	"github.com/gin-gonic/gin"
	"github.com/toolkits/pkg/ginx"
)

type hostCtfConfBody struct {
	LocalToml string `json:"local_toml"`
	IsApply   bool   `json:"is_apply"`
}

type multiHcRequestBody struct {
	Ips       []string `json:"ips"`
	LocalToml string   `json:"local_toml"`
	Name      string   `json:"name"`
	IsApply   bool     `json:"is_apply"`
}

type ipsRequestBody struct {
	Ips    []string `json:"ips"`
	Action string   `json:"action"`
}

type HostCtfRecord struct {
	Id           int64  `json:"id"`
	Name         string `json:"name"`
	Ip           string `json:"ip"`
	TemplateToml string `json:"template_toml"`
	LocalToml    string `json:"local_toml"`
	RemoteToml   string `json:"remote_toml"`
	Status       string `json:"status"`
	UpdateAt     int64  `json:"update_at"`
	UpdateBy     string `json:"update_by"`
}

func saveHostCtfConf(ip, name, localToml, remoteToml, status, msg, username string) error {
	// 保存/更新HostCtfConf

	var lastToml string
	var hc ltwmodels.HostCtfConf

	hc, err := ltwmodels.HostCtfConfGetByIpName(ip, name)
	ginx.Dangerous(err)

	if hc.Id == 0 {
		hc = ltwmodels.HostCtfConf{
			Ip:        ip,
			Name:      name,
			LocalToml: localToml,
			Status:    ltwmodels.HCStatus.INSTALLED,
			CreateBy:  username,
			UpdateBy:  username,
		}
		if localToml != "" {
			hc.LocalToml = localToml
		}
		if remoteToml != "" {
			hc.RemoteToml = remoteToml
		}
		if hc.RemoteToml != hc.LocalToml {
			hc.Status = ltwmodels.HCStatus.UNINSTALLED
		} else {
			hc.Status = ltwmodels.HCStatus.INSTALLED
		}

		hc, err = hc.Add()
	} else {
		lastToml = hc.RemoteToml
		hc.UpdateBy = username
		if localToml != "" {
			hc.LocalToml = localToml
		}
		if remoteToml != "" {
			hc.RemoteToml = remoteToml
		}
		if hc.RemoteToml != hc.LocalToml {
			hc.Status = ltwmodels.HCStatus.CONFLICTING
		} else {
			hc.Status = ltwmodels.HCStatus.INSTALLED
		}

		err = hc.Update(hc)
	}

	if err == nil {
		// 删除原先的索引数据
		hc.DelIndexById()
		// 创建新索引
		err = ctf.CreateQueryIndexes(hc)
		if err != nil {
			msg = msg + " 配置内容错误：" + err.Error()
		}
		if msg != "" {
			status = ltwmodels.CtfConfLogStatus.FAILED
			hc.Status = ltwmodels.HCStatus.ERROR
			err = errors.New(msg)
			hc.Update(hc)
		}
	}

	logErr := ltwmodels.AddHostCtfConfLog(
		hc.Id,
		hc.Ip,
		hc.Name,
		status,
		msg,
		lastToml,
		hc.RemoteToml,
		msg,
		hc.UpdateBy,
	)
	ginx.Dangerous(logErr)

	return err
}

func updateHostCtfConf(ip, name, toml, username string, isApply bool) error {
	var err error
	status := ltwmodels.CtfConfLogStatus.SUCCEED
	ospStdOut := ""
	remoteToml := toml

	if isApply {
		ospStdOut, err = ltw.UpdateCtfConf(ip, name, toml)
		if err != nil {
			ospStdOut = ospStdOut + err.Error()
			status = ltwmodels.CtfConfLogStatus.FAILED
			remoteToml = ""
		}
	}

	err = saveHostCtfConf(ip, name, toml, remoteToml, status, ospStdOut, username)

	return err
}

func CtfConfGets(c *gin.Context) {
	// 接口：获取监控项配置信息
	ginx.NewRender(c).Data(ctf.CatConfArr, nil)
}

func HostCtfConfGets(c *gin.Context) {
	// 接口：获取主机监控项列表
	ip := ginx.UrlParamStr(c, "ip")
	hcs, err := ltwmodels.GetHostCtfConfByIp(ip)
	var catRecord = map[string]ltwmodels.HostCtfConf{}
	for _, hc := range hcs {
		catRecord[hc.Name] = hc
	}

	var res []HostCtfRecord
	var tempConf HostCtfRecord
	for _, c := range ctf.CatConfArr {
		if value, isOk := catRecord[c.Name]; isOk {
			tempConf = HostCtfRecord{
				Name:         c.Name,
				TemplateToml: c.Toml,
				Id:           value.Id,
				Ip:           value.Ip,
				LocalToml:    value.LocalToml,
				RemoteToml:   value.RemoteToml,
				Status:       value.Status,
				UpdateBy:     value.UpdateBy,
				UpdateAt:     value.UpdateAt,
			}
		} else {
			tempConf = HostCtfRecord{
				Name:         c.Name,
				TemplateToml: c.Toml,
				Status:       ltwmodels.HCStatus.UNINSTALLED,
			}

		}
		res = append(res, tempConf)
	}
	ginx.Dangerous(err)

	ginx.NewRender(c).Data(gin.H{
		"list":  res,
		"total": len(res),
	}, nil)
}

func HostCtfConfPost(c *gin.Context) {
	// 接口：新增或修改监控项
	var b hostCtfConfBody
	ginx.BindJSON(c, &b)
	ip := ginx.UrlParamStr(c, "ip")

	err := updateHostCtfConf(
		ip,
		ginx.UrlParamStr(c, "name"),
		b.LocalToml,
		c.MustGet("user").(*models.User).Username,
		b.IsApply,
	)
	ginx.Dangerous(err)

	ginx.NewRender(c).Message(err)
}

func HostCtfConfDel(c *gin.Context) {
	// 接口：删除监控项
	id := ginx.UrlParamInt64(c, "id")
	logStatus := ltwmodels.CtfConfLogStatus.SUCCEED
	hc, err := ltwmodels.HostCtfConfGetById(id)
	lastToml := hc.RemoteToml
	ginx.Dangerous(err)

	// 1.删除远程配置文件
	if hc.Status == ltwmodels.HCStatus.UNINSTALLED {
		ginx.NewRender(c).Message("远程服务器未安装该监控，无须卸载！")
		return
	}
	msg, err := ltw.DisApplyConfig(hc.Ip, hc.Name)
	if err != nil {
		logStatus = ltwmodels.CtfConfLogStatus.FAILED
	}

	//2.删除本地配置数据
	hc.RemoteToml = ""
	hc.Status = ltwmodels.HCStatus.UNINSTALLED
	err = hc.Update(hc)
	ginx.Dangerous(err)

	// 3.添加操作日志
	logErr := ltwmodels.AddHostCtfConfLog(
		hc.Id,
		hc.Ip,
		hc.Name,
		logStatus,
		msg,
		lastToml,
		hc.RemoteToml,
		msg,
		hc.UpdateBy,
	)
	ginx.Dangerous(logErr)

	ginx.NewRender(c).Message(msg)
}

func HostCtfConfBatchPost(c *gin.Context) {
	// 接口：批量修改监控项
	var b multiHcRequestBody
	ginx.BindJSON(c, &b)
	username := c.MustGet("user").(*models.User).Username

	if len(b.Ips) == 1 {
		err := updateHostCtfConf(b.Ips[0], b.Name, b.LocalToml, username, b.IsApply)
		ginx.Dangerous(err)
	} else {
		wp := workerpool.New(10)
		for _, ip := range b.Ips {
			ip := ip
			wp.Submit(func() {
				updateHostCtfConf(ip, b.Name, b.LocalToml, username, b.IsApply)
			})
		}
	}

	ginx.NewRender(c).Message("")
}
