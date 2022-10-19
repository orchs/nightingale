package ltw

import (
	"context"
	"fmt"
	"github.com/didi/nightingale/v5/src/ltwmodels"
	"github.com/didi/nightingale/v5/src/models"
	"github.com/didi/nightingale/v5/src/pkg/ltw"
	"github.com/didi/nightingale/v5/src/storage"
	"github.com/didi/nightingale/v5/src/webapi/config"
	"github.com/gin-gonic/gin"
	nsema "github.com/niean/gotools/concurrent/semaphore"
	"github.com/toolkits/pkg/ginx"
	"net/http"
	"strings"
	"time"
)

const LastUpdateTargetId = "last_update_target_id"

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
			if ipMap[ch.Ip].Ip == "" {
				ch.Status = ltwmodels.HostCtfStatus.DISABLED
			} else {
				ch.Status = ipMap[ch.Ip].Status
				ch.Actions = strings.Split(ipMap[ch.Ip].Actions, ",")
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

func installLastVersion(ip, username string) (string, error) {
	status := ltwmodels.CtfConfLogStatus.SUCCEED
	var msg string

	std, err := ltw.InstallHostCtf(ip)
	if err != nil {
		msg = "安装失败！" + err.Error()
		status = ltwmodels.CtfConfLogStatus.FAILED
		ltwmodels.AddHostCtfConfLog(0, ip, "config", status, msg, "", "", std, username)
	} else {
		hc := ltwmodels.HostCtf{
			Ip:     ip,
			Status: ltwmodels.HostCtfStatus.ENABLED,
		}
		hc.Save()

		remoteToml, readErr := ltw.ReadConfToml(ip)
		if readErr != nil {
			msg += readErr.Error()
			ltwmodels.AddHostCtfConfLog(0, ip, "config", status, msg, "", "", std, username)
		} else {
			err = saveHostCtfConf(
				ip,
				"config",
				"",
				remoteToml,
				status,
				"",
				username,
			)
		}
	}

	return msg, err
}

func actionHostCtf(ip, script, sudoScript, action, status, actions, username string) (string, error) {
	logStatus := ltwmodels.CtfConfLogStatus.SUCCEED
	msg := "操作成功！"

	std, err := ltw.RunScript(ip, script, sudoScript)
	if err != nil {
		msg = "操作失败！" + err.Error()
		status = ltwmodels.CtfConfLogStatus.FAILED
		ltwmodels.AddHostCtfConfLog(0, ip, action, logStatus, msg, "", "", std, username)
		return msg, err
	}

	hc := ltwmodels.HostCtf{
		Ip:      ip,
		Status:  status,
		Actions: actions,
	}
	err = hc.Save()
	if err != nil {
		logStatus = ltwmodels.CtfConfLogStatus.FAILED
		msg = "安装成功！更新状态失败！"
	}
	ltwmodels.AddHostCtfConfLog(0, ip, action, logStatus, msg, "", "", std, username)

	return msg, err
}

func uninstallCtf(ip, username string) (string, error) {
	var msg string
	status := ltwmodels.CtfConfLogStatus.SUCCEED
	std, err := ltw.StopHostCtf(ip)
	if err != nil {
		status = ltwmodels.CtfConfLogStatus.FAILED
		msg = "卸载失败！" + msg
	} else {
		hc := ltwmodels.HostCtf{
			Ip:     ip,
			Status: ltwmodels.HostCtfStatus.DISABLED,
		}
		hc.Save()
		msg = "卸载成功！"
	}
	ltwmodels.AddHostCtfConfLog(0, ip, "config", status, msg, "", "", std, username)

	return std, err
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

func getPullScriptScript() (string, string) {
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
	var newActions string
	if b.Action == "INSTALL" {
		script, sudoScript = getInstallScript()
		actionName = "安装categraf"
		newStatus = "ENABLED"
		newActions = "DISABLE,UNINSTALL"
	} else if b.Action == "UNINSTALL" {
		script, sudoScript = getUnInstallScript()
		actionName = "卸载categraf"
		newStatus = "UNINSTALLED"
		newActions = "INSTALL"
	} else if b.Action == "ENABLE" {
		script = "systemctl start categraf"
		sudoScript = "sudo systemctl start categraf"
		actionName = "启用categraf"
		newStatus = "ENABLED"
		newActions = "ENABLED,UNINSTALL"
	} else if b.Action == "DISABLE" {
		script = "systemctl stop categraf"
		sudoScript = "sudo systemctl stop categraf"
		actionName = "禁用categraf"
		newStatus = "DISABLED"
		newActions = "ENABLED,UNINSTALL"
	} else if b.Action == "PULL_SCRIPTS" {
		script, sudoScript = getPullScriptScript()
		actionName = "更新exec脚本"
	}

	if len(b.Ips) == 1 {
		_, err := actionHostCtf(b.Ips[0], script, sudoScript, actionName, newStatus, newActions, username)
		ginx.Dangerous(err)
	} else {
		concurrentNum := 100
		sema := nsema.NewSemaphore(concurrentNum)

		for _, ip := range b.Ips {
			go func(ip, script, sudoScript, actionName, newStatus, newActions, username string) {
				if !sema.TryAcquire() {
					return
				}
				defer sema.Release()

				actionHostCtf(ip, script, sudoScript, actionName, newStatus, newActions, username)
			}(ip, script, sudoScript, actionName, newStatus, newActions, username)
		}
	}

	ginx.NewRender(c).Message("")
}

func HostCtfPost(c *gin.Context) {
	// 接口：批量安装categraf
	var b ipsRequestBody
	ginx.BindJSON(c, &b)
	username := c.MustGet("user").(*models.User).Username

	concurrentNum := 100
	sema := nsema.NewSemaphore(concurrentNum)

	if len(b.Ips) == 1 {
		_, err := installLastVersion(b.Ips[0], username)
		ginx.Dangerous(err)
	} else {
		for _, ip := range b.Ips {
			go func(ip, username string) {
				if !sema.TryAcquire() {
					return
				}
				defer sema.Release()

				installLastVersion(ip, username)
			}(ip, username)
		}
	}

	ginx.NewRender(c).Message("")
}

func HostCtfDelete(c *gin.Context) {
	// 接口：批量卸载categraf
	var b ipsRequestBody
	ginx.BindJSON(c, &b)
	username := c.MustGet("user").(*models.User).Username

	if len(b.Ips) == 1 {
		_, err := uninstallCtf(b.Ips[0], username)
		ginx.Dangerous(err)
	} else {
		for _, ip := range b.Ips {
			go uninstallCtf(ip, username)
		}
	}
	ginx.NewRender(c).Message("")
}

func autoAddCtfConf() {
	ctx := context.Background()

	id, err := storage.Redis.Get(ctx, LastUpdateTargetId).Int64()
	if err != nil {
		id = 1
	}

	list, err := models.LtwIdentGets(id)
	if err != nil {
		return
	}

	for _, ident := range list {
		hostname, _, err := ltw.SplitHostnameAndIp(ident.Ident)

		chs, err := ltw.GetCmdbHosts(1, "root", hostname)
		if err != nil {
			continue
		}
		hc := ltwmodels.HostCtf{
			Ip:     chs[0].Ip,
			Status: ltwmodels.HostCtfStatus.ENABLED,
		}
		hc.Save()
		id = ident.Id
	}

	storage.Redis.Set(ctx, LastUpdateTargetId, id, time.Duration(time.Hour*24)).Err()
}

func UpdateCtfStatus() {
	for {
		time.Sleep(time.Second * 15)
		autoAddCtfConf()
	}
}
