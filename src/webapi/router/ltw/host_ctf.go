package ltw

import (
	"fmt"
	"github.com/didi/nightingale/v5/src/ltwmodels"
	"github.com/didi/nightingale/v5/src/models"
	"github.com/didi/nightingale/v5/src/pkg/ltw"
	"github.com/didi/nightingale/v5/src/webapi/config"
	"github.com/gammazero/workerpool"
	"github.com/gin-gonic/gin"
	"github.com/toolkits/pkg/ginx"
	"net/http"
	"time"
)

func HostCtfGets(c *gin.Context) {
	// 接口：根据主机名、监控项查询主机列表
	tag := ginx.QueryInt64(c, "tag", 1)
	query := ginx.QueryStr(c, "query", "")
	name := ginx.QueryStr(c, "name", "")
	index := ginx.QueryStr(c, "key", "")
	value := ginx.QueryStr(c, "value", "")
	var ipStr string
	ipMap := make(map[string]ltwmodels.HostCtf)
	user := c.MustGet("user").(*models.User).Username
	statusFlag := false

	if query == "" && name == "" {
		ginx.NewRender(c, http.StatusBadRequest).Message("请至少选择一个查询条件！")
		return
	}

	// 1.查看是否根据指标项搜索
	if name != "" {
		hosts, err := ltwmodels.GetHostCtfByConf(name, index, value)
		ginx.Dangerous(err)
		for _, host := range hosts {
			ipStr = ipStr + " " + host.Ip
			ipMap[host.Ip] = host
		}
		statusFlag = true
	}

	// 2.查询cmdb主机信息
	var cmdbHosts []ltwmodels.CmdbHostInfo
	var err error
	if query == "" {
		if ipStr != "" {
			cmdbHosts, err = ltw.GetCmdbHosts(tag, user, ipStr)
		}
	} else {
		cmdbHosts, err = ltw.GetCmdbHosts(tag, user, query)
	}

	var resHosts []ltwmodels.CmdbHostInfo
	if statusFlag {
		// 取cmdb和监控项的交集
		for _, ch := range cmdbHosts {
			if v, ok := ipMap[ch.Ip]; ok {
				ch.Status = v.Status
				ch.Version = v.Version
				if ch.Status == ltwmodels.HostCtfStatus.ENABLED {
					if ch.Version == config.C.Ltw.CtfVersion {
						ch.Actions = []string{"DISABLE"}
					} else {
						ch.Actions = []string{"DISABLE", "UPDATE"}
					}
				} else if ch.Status == ltwmodels.HostCtfStatus.DISABLED {
					if ch.Version == config.C.Ltw.CtfVersion {
						ch.Actions = []string{"ENABLE"}
					} else {
						ch.Actions = []string{"ENABLE", "UPDATE"}
					}
				} else {
					ch.Status = ltwmodels.HostCtfStatus.UNINSTALLED
					ch.Actions = []string{"INSTALL"}
				}
				resHosts = append(resHosts, ch)
			}
		}
	} else {
		// 直接取cmdb的内容
		var ips []string
		for _, ch := range cmdbHosts {
			ips = append(ips, ch.Ip)
		}
		hosts, err := ltwmodels.GetHostCtfByIps(ips)
		ginx.Dangerous(err)

		for _, host := range hosts {
			ipMap[host.Ip] = host
		}

		for _, ch := range cmdbHosts {
			ch.Status = ipMap[ch.Ip].Status
			ch.Version = ipMap[ch.Ip].Version
			if ch.Status == ltwmodels.HostCtfStatus.ENABLED {
				if ch.Version == config.C.Ltw.CtfVersion {
					ch.Actions = []string{"DISABLE"}
				} else {
					ch.Actions = []string{"DISABLE", "UPDATE"}
				}
			} else if ch.Status == ltwmodels.HostCtfStatus.DISABLED {
				if ch.Version == config.C.Ltw.CtfVersion {
					ch.Actions = []string{"ENABLE"}
				} else {
					ch.Actions = []string{"ENABLE", "UPDATE"}
				}
			} else {
				ch.Status = ltwmodels.HostCtfStatus.UNINSTALLED
				ch.Actions = []string{"INSTALL"}
			}
			resHosts = append(resHosts, ch)
		}
	}

	go ltw.StorageHostToRedis(resHosts)

	ginx.Dangerous(err)
	ginx.NewRender(c).Data(gin.H{
		"list": resHosts,
	}, nil)

}

func actionHostCtf(ip, script, sudoScript, action, status, username string) (string, error) {
	logStatus := ltwmodels.CtfConfLogStatus.SUCCEED
	msg := "操作成功！"

	std, err := ltw.RunScript(ip, script, sudoScript)
	if err != nil {
		msg = "操作失败！" + err.Error()
		logStatus := ltwmodels.CtfConfLogStatus.FAILED
		ltwmodels.AddHostCtfConfLog(0, ip, action, logStatus, msg, "", "", std, username)
		return msg, err
	}

	hc := ltwmodels.HostCtf{
		Ip:      ip,
		Status:  status,
		Version: config.C.Ltw.CtfVersion,
	}
	err = hc.Save()
	if err != nil {
		logStatus = ltwmodels.CtfConfLogStatus.FAILED
		msg = "安装成功！更新状态失败！"
	}
	ltwmodels.AddHostCtfConfLog(0, ip, action, logStatus, msg, "", "", std, username)

	return msg, err
}

func getInstallScript() (string, string) {
	var script string
	var sudoScript string

	// 清理、备份历史脚本
	cs := fmt.Sprintf(
		"if [ -f /tmp/install_ctf.sh ];then mv /tmp/install_ctf.sh /tmp/install_ctf_%v.sh;fi",
		time.Now().Unix(),
	)

	// 下载新脚本
	ds := fmt.Sprintf(
		"wget -q --no-check-certificate %v/install_ctf.sh --http-user=%v --http-password=%v -P /tmp/ ",
		config.C.Ltw.CtfPkgDownloadPath,
		config.C.Ltw.CtfPkgDownloadUser,
		config.C.Ltw.CtfPkgDownloadPass,
	)

	// 执行脚本
	is := "bash /tmp/install_ctf.sh \n"
	script = cs + " && " + ds + " && " + is
	sudoScript = cs + " && " + ds + " && sudo " + is

	return script, sudoScript
}

func getUnInstallScript() (string, string) {
	var script string
	var sudoScript string

	// 清理、备份历史脚本
	cs := fmt.Sprintf(
		"if [ -f /tmp/uninstall_ctf.sh ];then mv /tmp/uninstall_ctf.sh /tmp/uninstall_ctf_%v.sh;fi",
		time.Now().Unix(),
	)

	// 下载新脚本
	ds := fmt.Sprintf(
		"wget -q --no-check-certificate %v/uninstall_ctf.sh --http-user=%v --http-password=%v -P /tmp/ ",
		config.C.Ltw.CtfPkgDownloadPath,
		config.C.Ltw.CtfPkgDownloadUser,
		config.C.Ltw.CtfPkgDownloadPass,
	)

	// 执行脚本
	is := "bash /tmp/uninstall_ctf.sh \n"
	script = cs + " && " + ds + " && " + is
	sudoScript = cs + " && " + ds + " && sudo " + is

	return script, sudoScript
}

func getUpdateScriptScript() (string, string) {
	var script string
	var sudoScript string

	// 清理、备份历史脚本
	cs := fmt.Sprintf(
		"if [ -f /tmp/pull_scripts.sh ];then mv /tmp/pull_scripts.sh /tmp/pull_scripts%v.sh;fi",
		time.Now().Unix(),
	)

	// 下载新脚本
	ds := fmt.Sprintf(
		"wget -q --no-check-certificate %v/pull_scripts.sh --http-user=%v --http-password=%v -P /tmp/ ",
		config.C.Ltw.CtfPkgDownloadPath,
		config.C.Ltw.CtfPkgDownloadUser,
		config.C.Ltw.CtfPkgDownloadPass,
	)

	// 执行脚本
	is := "bash /tmp/pull_scripts.sh \n"
	script = cs + " && " + ds + " && " + is
	sudoScript = cs + " && " + ds + " && sudo " + is

	return script, sudoScript
}

func HostCtfPostNew(c *gin.Context) {
	// 接口：批量安装categraf
	var b ipsRequestBody
	ginx.BindJSON(c, &b)
	username := c.MustGet("user").(*models.User).Username

	var script string
	var sudoScript string
	var actionName string
	var newStatus string
	if b.Action == "INSTALL" {
		script, sudoScript = getInstallScript()
		actionName = "安装categraf"
		newStatus = "ENABLED"
	} else if b.Action == "UNINSTALL" {
		script, sudoScript = getUnInstallScript()
		actionName = "卸载categraf"
		newStatus = "UNINSTALLED"
	} else if b.Action == "ENABLE" {
		script = "systemctl start categraf"
		sudoScript = "sudo systemctl start categraf"
		actionName = "启用categraf"
		newStatus = "ENABLED"
	} else if b.Action == "DISABLE" {
		script = "systemctl stop categraf"
		sudoScript = "sudo systemctl stop categraf"
		actionName = "禁用categraf"
		newStatus = "DISABLED"
	} else if b.Action == "UPDATE" {
		script, sudoScript = getUpdateScriptScript()
		actionName = "更新categraf"
		newStatus = "ENABLED"
	}

	if len(b.Ips) == 1 {
		_, err := actionHostCtf(b.Ips[0], script, sudoScript, actionName, newStatus, username)
		ginx.Dangerous(err)
	} else {
		wp := workerpool.New(10)
		for _, ip := range b.Ips {
			ip := ip
			wp.Submit(func() {
				actionHostCtf(ip, script, sudoScript, actionName, newStatus, username)
			})
		}
	}

	ginx.NewRender(c).Message("")
}