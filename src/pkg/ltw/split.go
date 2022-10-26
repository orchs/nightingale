package ltw

import (
	"errors"
	"strings"
)

func SplitHostnameAndIp(t string) (string, string, error) {
	a := strings.Split(t, "_")
	s := make([]string, len(a)-1)
	copy(s, a[:])
	return strings.Join(s, "_"), a[len(a)-1], nil
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
