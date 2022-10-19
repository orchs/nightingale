package ltw

import (
	"errors"
	"strings"
)

func SplitHostnameAndIp(t string) (string, string, error) {
	s := strings.Split(t, "_")
	if len(s) != 2 {
		return "", "", errors.New("target格式错误！")
	}
	return s[0], s[1], nil
}

func ResolveAdminInfo(info string) (string, string, error) {
	t := strings.Split(info, ":")
	if len(t) != 2 {
		return "", "", errors.New("主机管理员信息错误，请在cmdb中进行查看！")
	}
	rsa, err := GetSecret(t[0], t[1])
	if err != nil {
		return "", "", errors.New("秘钥获取失败，请确认秘钥是否覆盖！")
	}

	return strings.Split(t[0], "_")[1], rsa, nil
}
