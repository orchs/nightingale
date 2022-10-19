// ltw router

package ltw

import (
	"github.com/didi/nightingale/v5/src/ltwmodels"
	"github.com/didi/nightingale/v5/src/models"
	"github.com/didi/nightingale/v5/src/pkg/ltw"
	"github.com/didi/nightingale/v5/src/pkg/osp"
	"github.com/gin-gonic/gin"
	nsema "github.com/niean/gotools/concurrent/semaphore"
	"github.com/toolkits/pkg/ginx"
	"net/http"
	"strings"
)

var LogSize = 500

func HostGets(c *gin.Context) {
	tag := ginx.QueryInt64(c, "tag", 1)
	query := ginx.QueryStr(c, "query", "")
	user := c.MustGet("user").(*models.User).Username

	if query == "" {
		ginx.NewRender(c, http.StatusBadRequest).Message("查询条件不能为空！")
		return
	}

	// 获取target列表
	ts, err := models.TargetGets(-1, nil, query, 10000, 0)
	ginx.Dangerous(err)
	tm := make(map[string]bool)
	for _, t := range ts {
		hostName := strings.Split(t.Ident, "_")[0]
		tm[hostName] = true
	}

	// 获取cmdb的权限主机列表
	chs, err := ltw.GetCmdbHosts(tag, user, query)
	var list []ltwmodels.HostInfo
	for _, ch := range chs {
		//if tm[ch.HostName] {
		list = append(list, ltwmodels.HostInfo{
			Ip:        ch.Ip,
			HostName:  ch.HostName,
			Port:      ch.Port,
			AdminUser: ch.AdminUser,
		})
		//}
	}

	ginx.NewRender(c).Data(gin.H{
		"list":  list,
		"total": len(list),
	}, nil)
}

func HostCtfPosts(c *gin.Context) {
	var ips []string
	ginx.BindJSON(c, &ips)
	username := c.MustGet("user").(*models.User).Username

	script, sudoScript := getInstallScript()

	concurrentNum := 100
	sema := nsema.NewSemaphore(concurrentNum)

	for _, ip := range ips {
		go func(ip, username, script, sudoScript string) {
			if !sema.TryAcquire() {
				return
			}
			defer sema.Release()
			installCtf(ip, username, script, sudoScript)
		}(ip, username, script, sudoScript)
	}

	ginx.NewRender(c).Message("")
}

func installCtf(ip, username, script, sudoScript string) {
	chs, err := ltw.GetCmdbHosts(1, username, ip)
	if err != nil {
		return
	}
	if len(chs) < 0 {
		return
	}

	h := chs[0]

	user := h.AdminUser
	if h.Rsa == "" {
		t := strings.Split(h.AdminUser, ":")
		if len(t) < 2 {
			return
		}
		rsa, err := ltw.GetSecret(t[0], t[1])
		if err != nil {
			return
		}
		h.Rsa = rsa

		user = strings.Split(t[0], "_")[1]
	}

	rs := script
	if user != "root" {
		rs = sudoScript
	}

	std, err := osp.ExecScript(h.Ip, h.Port, user, h.Passwd, h.Rsa, rs)
	stats := ltwmodels.CtfLogStatusChoose.SUCCEED
	if err != nil {
		stats = ltwmodels.CtfLogStatusChoose.FAILED
	}
	if len(std) > LogSize {
		std = string([]byte(std)[:LogSize])
	}

	// 记录安装日志
	l := ltwmodels.CtfLogs{
		Ip:       h.Ip,
		Hostname: h.HostName,
		Action:   ltwmodels.CtfLogActionChoose.INSTALL,
		Status:   stats,
		StandOut: std,
	}
	l.Add()
}
