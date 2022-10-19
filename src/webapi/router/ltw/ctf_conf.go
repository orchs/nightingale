package ltw

import (
	"errors"
	"fmt"
	"github.com/didi/nightingale/v5/src/ltwmodels"
	"github.com/didi/nightingale/v5/src/models"
	"github.com/didi/nightingale/v5/src/pkg/ltw"
	"github.com/didi/nightingale/v5/src/pkg/ltw/ctf"
	"github.com/didi/nightingale/v5/src/pkg/osp"
	"github.com/gin-gonic/gin"
	nsema "github.com/niean/gotools/concurrent/semaphore"
	"github.com/toolkits/pkg/ginx"
	"net/http"
	"strings"
)

type TargetConfForm struct {
	Idents []string `json:"idents" binding:"required"`
	Name   string   `json:"name"`
	Toml   string   `json:"Toml"`
}

const (
	restartCtf     = "systemctl restart categraf\n"
	sudoRestartCtf = "sudo systemctl restart categraf\n"
	stopCtf        = "systemctl stop categraf\n"
	sudoStopCtf    = "sudo systemctl stop categraf\n"
)

func TargetConfPosts(c *gin.Context) {
	username := c.MustGet("user").(*models.User).Username
	var f TargetConfForm
	ginx.BindJSON(c, &f)

	if len(f.Idents) == 0 {
		ginx.Bomb(http.StatusBadRequest, "idents empty")
	}

	if len(f.Toml) == 0 {
		ginx.Bomb(http.StatusBadRequest, "toml empty")
	}
	script, sudoScript := getUpdateConfScript(f.Name, f.Toml)

	concurrentNum := 100
	sema := nsema.NewSemaphore(concurrentNum)
	for _, ident := range f.Idents {
		go func(ident, script, sudoScript, username, toml string) {
			if !sema.TryAcquire() {
				return
			}
			defer sema.Release()
			updateCtfConf(ident, script, sudoScript, username, toml)
		}(ident, script, sudoScript, username, f.Toml)
	}
	ginx.NewRender(c).Message("")
}

func getUpdateConfScript(name, toml string) (string, string) {
	dirPath := ctf.GetConfigDirPathByName(name)
	mkdirScript := fmt.Sprintf("mkdir -p %v && chmod 777 %v \n", dirPath, dirPath)
	confPath := ctf.GetConfigFilePathByName(name)
	editScript := fmt.Sprintf(`
cat > %v << EOF 
%v 
EOF
`, confPath, toml)

	if name == "config" {
		editScript = strings.Replace(editScript, "$hostname", "\\$hostname", 10)
		editScript = strings.Replace(editScript, "$ip", "\\$ip", 10)
	}

	script := mkdirScript + editScript + restartCtf
	sudoScript := "sudo " + mkdirScript + "sudo " + editScript + sudoRestartCtf
	return script, sudoScript
}
func getHostInfo(ident, username string) (string, string, string, string, string, string, error) {
	hostname, ip, err := ltw.SplitHostnameAndIp(ident)
	if err != nil {
		return "", "", "", "", "", "", err
	}

	chs, err := ltw.GetCmdbHosts(1, username, ip)
	if err != nil {
		return "", "", "", "", "", "", errors.New(fmt.Sprintf("在cmdb中查询到多个：%v", err))
	} else if len(chs) != 1 {
		return "", "", "", "", "", "", errors.New("在cmdb中查询到多个：" + ip)
	}

	h := chs[0]
	h.User, h.Rsa, err = ltw.ResolveAdminInfo(h.AdminUser)
	if err != nil {
		return "", "", "", "", "", "", errors.New(fmt.Sprintf("解析主机管理员信息出错：%v", err))
	} else {
	}
	return hostname, h.Ip, h.Port, h.User, h.Passwd, h.Rsa, nil
}
func updateCtfConf(ident, script, sudoScript, username, toml string) {
	stats := ltwmodels.CtfItemLogStatusChoose.FAILED
	var std string
	hostname, ip, port, user, passwd, rsa, err := getHostInfo(ident, username)
	if err != nil {
		std = fmt.Sprintf("%v", err)
	} else {
		rs := script
		if user != "root" {
			rs = sudoScript
		}

		std, err = osp.ExecScript(ip, port, user, passwd, rsa, rs)

		if err != nil {
			std = fmt.Sprintf("下发配置信息出错：%v", err)
		} else {
			stats = ltwmodels.CtfItemLogStatusChoose.SUCCEED
			if len(std) > LogSize {
				std = string([]byte(std)[:LogSize])
			}
		}
	}

	l := ltwmodels.CtfItemLogs{
		Ident:    ident,
		Ip:       ip,
		Hostname: hostname,
		Action:   ltwmodels.CtfItemLogActionChoose.UPDATE,
		Status:   stats,
		StandOut: std,
		Toml:     toml,
		CreateBy: username,
	}
	l.Add()
}

func TargetConfGets(c *gin.Context) {
	username := c.MustGet("user").(*models.User).Username
	target := ginx.UrlParamStr(c, "target")
	_, ip, err := ltw.SplitHostnameAndIp(target)
	if err != nil {
		return
	}

	chs, err := ltw.GetCmdbHosts(1, username, ip)
	if err != nil {
		ginx.Bomb(http.StatusBadRequest, "出错：查询主机信息出错，ip：%v！", ip)
	}
	if len(chs) != 1 {
		ginx.Bomb(http.StatusBadRequest, "出错：%v在cmdb中查询到多个，请确认！", ip)
	}

	h := chs[0]
	h.User, h.Rsa, err = ltw.ResolveAdminInfo(h.AdminUser)
	if err != nil {
		ginx.Bomb(http.StatusBadRequest, "解析%v的管理员信息出错：%v！", ip, err)

	}

	script := fmt.Sprintf("/opt/categraf/get_conf.shell")
	std, err := osp.ExecScript(h.Ip, h.Port, h.User, h.Passwd, h.Rsa, script)

	ctfMap := make(map[string]string)
	for _, item := range strings.Split(std, "-----") {
		s := strings.Split(item, "|||||")
		ctfMap[s[0]] = s[1]
	}

	var lst []ctf.ItemInfo
	for _, i := range ctf.Items {
		if v, isOk := ctfMap[i.Name]; isOk {
			i.Toml = v
			i.Status = true
		}
		lst = append(lst, i)
	}

	ginx.NewRender(c).Data(gin.H{
		"items": lst,
	}, nil)
}

func TargetConfGet(c *gin.Context) {
	username := c.MustGet("user").(*models.User).Username
	target := ginx.UrlParamStr(c, "target")
	ctfItem := ginx.UrlParamStr(c, "ctf")
	_, ip, err := ltw.SplitHostnameAndIp(target)
	if err != nil {
		return
	}

	chs, err := ltw.GetCmdbHosts(1, username, ip)
	if err != nil {
		ginx.Bomb(http.StatusBadRequest, "出错：查询主机信息出错，ip：%v！", ip)
	}
	if len(chs) != 1 {
		ginx.Bomb(http.StatusBadRequest, "出错：%v在cmdb中查询到多个，请确认！", ip)
	}

	h := chs[0]
	h.User, h.Rsa, err = ltw.ResolveAdminInfo(h.AdminUser)
	if err != nil {
		ginx.Bomb(http.StatusBadRequest, "解析%v的管理员信息出错：%v！", ip, err)
	}
	script := fmt.Sprintf("cat /opt/categraf/conf/input.%v/%v.toml", ctfItem, ctfItem)
	std, err := osp.ExecScript(h.Ip, h.Port, h.User, h.Passwd, h.Rsa, script)
	if err != nil {
		ginx.NewRender(c).Data(gin.H{
			"remote_toml": "拿模板内容",
			"status":      "未安装",
		}, nil)
	} else {
		ginx.NewRender(c).Data(gin.H{
			"remote_toml": std,
			"status":      "已安装",
		}, nil)
	}
}

func targetCtfSave(t, toml string) {

}
