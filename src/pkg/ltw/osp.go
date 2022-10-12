package ltw

import (
	"context"
	"fmt"
	"github.com/didi/nightingale/v5/src/ltwmodels"
	"github.com/didi/nightingale/v5/src/pkg/ltw/ctf"
	"github.com/didi/nightingale/v5/src/storage"
	"github.com/didi/nightingale/v5/src/webapi/config"
	"github.com/mitchellh/mapstructure"
	gossh "golang.org/x/crypto/ssh"
	"net"
	"strings"
	"time"
)

const (
	restartCtf     = "systemctl restart categraf\n"
	sudoRestartCtf = "sudo systemctl restart categraf\n"
	stopCtf        = "systemctl stop categraf\n"
	sudoStopCtf    = "sudo systemctl stop categraf\n"
)

type Cli struct {
	ip         string
	client     *gossh.Client
	session    *gossh.Session
	LastResult string
}

func in(target string, s []string) bool {
	for _, element := range s {
		if target == element {
			return true
		}
	}
	return false
}

func (c *Cli) Connect() (*Cli, string, error) {

	// 连接远程服务器
	ctx := context.Background()
	host, err := storage.Redis.HGetAll(ctx, "host_"+c.ip).Result()
	h := ltwmodels.CmdbHostInfo{}
	err = mapstructure.Decode(host, &h)
	if err != nil {
		return c, "", err
	}

	var signer gossh.Signer
	if config.C.Ltw.HostSSHType == "RSA" {
		signer, err = gossh.ParsePrivateKey([]byte(h.Rsa))
	} else {
		signer, err = gossh.ParsePrivateKeyWithPassphrase([]byte(h.Rsa), []byte(h.Passwd))
	}
	if err != nil {
		return c, "", err
	}

	config := &gossh.ClientConfig{}
	config.SetDefaults()
	config.Timeout = time.Second
	config.User = h.User
	config.Auth = []gossh.AuthMethod{gossh.PublicKeys(signer)}
	config.HostKeyCallback = func(hostname string, remote net.Addr, key gossh.PublicKey) error { return nil }
	client, err := gossh.Dial("tcp", c.ip+":"+h.Port, config)
	if nil != err {
		return c, "", err
	}
	c.client = client

	return c, h.User, nil
}

func (c Cli) Run(shell string) (string, error) {
	// 执行shell
	if c.client == nil {
		if _, _, err := c.Connect(); err != nil {
			return "", err
		}
	}

	session, err := c.client.NewSession()
	if err != nil {
		return "", err
	}

	// 关闭会话
	defer session.Close()
	buf, err := session.CombinedOutput(shell)

	c.LastResult = string(buf)
	return c.LastResult, err
}

func UpdateCtfConf(ip, confName, toml string) (string, error) {
	// 更新ctf配置文件内容
	cli := Cli{
		ip: ip,
	}

	c, username, err := cli.Connect()
	if err != nil {
		return "", err
	}
	defer c.client.Close()

	dirPath := ctf.GetConfigDirPathByName(confName)
	mkdirScript := fmt.Sprintf("mkdir -p %v && chmod 777 %v \n", dirPath, dirPath)
	confPath := ctf.GetConfigFilePathByName(confName)
	editConfigScript := fmt.Sprintf(`
cat > %v << EOF 
%v 
EOF
`, confPath, toml)
	restartScript := restartCtf
	if username != "root" {
		mkdirScript = fmt.Sprintf("sudo mkdir -p %v && sudo chmod 777 %v \n", dirPath, dirPath)
		editConfigScript = fmt.Sprintf(`
sudo cat > %v << EOF 
%v 
EOF
`, confPath, toml)
		restartScript = sudoRestartCtf
	}

	if confName == "config" {
		editConfigScript = strings.Replace(editConfigScript, "$hostname", "\\$hostname", 10)
		editConfigScript = strings.Replace(editConfigScript, "$ip", "\\$ip", 10)
	}

	shellScript := mkdirScript + editConfigScript + restartScript
	return c.Run(shellScript)
}

func DisApplyConfig(ip, confName string) (string, error) {
	// 禁用ctf配置
	cli := Cli{
		ip: ip,
	}

	c, username, err := cli.Connect()
	if err != nil {
		return "连接目标机器失败！", err
	}
	defer c.client.Close()

	dirPath := ctf.GetConfigDirPathByName(confName)
	batDirPath := ctf.GetBatConfigDirPathByName(confName)

	shellScript := stopCtf
	if username != "root" {
		shellScript = sudoStopCtf
	}

	if confName != "config" {
		shellScript = fmt.Sprintf("mv %v %v", dirPath, batDirPath)
		shellScript += " && " + restartCtf
		if username != "root" {
			shellScript = fmt.Sprintf("sudo mv %v %v", dirPath, batDirPath)
			shellScript += " && " + sudoRestartCtf
		}
	}

	return c.Run(shellScript)
}

func InstallHostCtf(ip string) (string, error) {
	// 安装categraf
	cli := Cli{ip: ip}

	c, username, err := cli.Connect()
	if err != nil {
		return "连接目标机器失败！", err
	}
	defer c.client.Close()

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

	// 安装脚本
	ss := "bash /tmp/install_ctf.sh \n"
	if username != "root" {
		ss = "sudo bash /tmp/install_ctf.sh \n"
	}
	shellScript := cs + " && " + ds + " && " + ss

	return c.Run(shellScript)
}

func ReadConfToml(ip string) (string, error) {
	// 禁用ctf配置
	cli := Cli{
		ip: ip,
	}

	c, _, err := cli.Connect()
	if err != nil {
		return "连接目标机器失败！", err
	}
	defer c.client.Close()

	shellScript := `
cat /opt/categraf/conf/config.toml
`
	return c.Run(shellScript)
}

func StopHostCtf(ip string) (string, error) {
	// 禁用ctf配置
	cli := Cli{
		ip: ip,
	}

	c, username, err := cli.Connect()
	if err != nil {
		return "连接目标机器失败！", err
	}
	defer c.client.Close()

	shellScript := stopCtf
	if username != "root" {
		shellScript = sudoStopCtf
	}

	return c.Run(shellScript)
}

func RunScript(ip, script, sudoScript string) (string, error) {
	cli := Cli{
		ip: ip,
	}
	c, user, err := cli.Connect()
	if err != nil {
		return "连接目标机器失败！", err
	}

	session, err := c.client.NewSession()
	if err != nil {
		return "", err
	}

	shell := script
	if user != "root" {
		shell = sudoScript
	}
	// 关闭会话
	defer session.Close()
	buf, err := session.CombinedOutput(shell)

	c.LastResult = string(buf)
	return c.LastResult, err
}
