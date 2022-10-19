package osp

import (
	"github.com/didi/nightingale/v5/src/webapi/config"
	gossh "golang.org/x/crypto/ssh"
	"net"
	"time"
)

const (
	restartCtf     = "systemctl restart categraf\n"
	sudoRestartCtf = "sudo systemctl restart categraf\n"
	stopCtf        = "systemctl stop categraf\n"
	sudoStopCtf    = "sudo systemctl stop categraf\n"
)

type Cli struct {
	Ip         string
	Port       string
	User       string
	Password   string
	Rsa        string
	Script     string
	Client     *gossh.Client
	session    *gossh.Session
	LastResult string
}

func (c *Cli) Connect() (*Cli, error) {
	var signer gossh.Signer
	var err error
	if config.C.Ltw.HostSSHType == "RSA" {
		signer, err = gossh.ParsePrivateKey([]byte(c.Rsa))
	} else {
		signer, err = gossh.ParsePrivateKeyWithPassphrase([]byte(c.Rsa), []byte(c.Password))
	}
	if err != nil {
		return c, err
	}

	config := &gossh.ClientConfig{}
	config.SetDefaults()
	config.Timeout = time.Second
	config.User = c.User
	config.Auth = []gossh.AuthMethod{gossh.PublicKeys(signer)}
	config.HostKeyCallback = func(hostname string, remote net.Addr, key gossh.PublicKey) error { return nil }
	client, err := gossh.Dial("tcp", c.Ip+":"+c.Port, config)
	if nil != err {
		return c, err
	}
	c.Client = client

	return c, nil
}

func ExecScript(ip, port, user, password, rsa, script string) (string, error) {

	cli := Cli{
		Ip:       ip,
		Port:     port,
		User:     user,
		Password: password,
		Rsa:      rsa,
		Script:   script,
	}
	c, err := cli.Connect()
	if err != nil {
		return "连接目标机器失败！", err
	}

	session, err := c.Client.NewSession()
	if err != nil {
		return "", err
	}

	// 关闭会话
	defer session.Close()
	buf, err := session.CombinedOutput(script)

	c.LastResult = string(buf)
	return c.LastResult, err
}
