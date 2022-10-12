package ltw

import (
	"fmt"
	"github.com/didi/nightingale/v5/src/ltwmodels"
	"github.com/didi/nightingale/v5/src/models"
	"strings"

	"context"
	"encoding/json"
	"errors"
	"github.com/didi/nightingale/v5/src/storage"
	"github.com/didi/nightingale/v5/src/webapi/config"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)

type CmdbHostResponse struct {
	Code  int                      `json:"code"`
	Msg   string                   `json:"msg"`
	Count int                      `json:"count"`
	Data  []ltwmodels.CmdbHostInfo `json:"data"`
}

//Get http get method
func Get(url string, params map[string]string, headers map[string]string) ([]byte, error) {
	//new request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Println(err)
		return nil, errors.New("new request is fail ")
	}
	//add params
	q := req.URL.Query()
	if params != nil {
		for key, val := range params {
			q.Add(key, val)
		}
		req.URL.RawQuery = q.Encode()
	}
	//add headers
	if headers != nil {
		for key, val := range headers {
			req.Header.Add(key, val)
		}
	}
	//http client
	client := &http.Client{}
	log.Printf("Go GET URL : %s \n", req.URL.String())

	//发送请求
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close() //一定要关闭res.Body
	//读取body
	resBody, err := ioutil.ReadAll(res.Body) //把  body 内容读入字符串 s
	if err != nil {
		return nil, err
	}
	return resBody, nil
}

func GetCmdbHosts(user *models.User, tag int64, query string) ([]ltwmodels.CmdbHostInfo, error) {
	// 获取cmdb主机列表

	url := config.C.Ltw.HostQueryUrl
	params := make(map[string]string)
	params["auth_key"] = config.C.Ltw.HostQueryKey
	params["tag"] = strconv.FormatInt(int64(tag), 10)

	if user.Username != "root" {
		params["user"] = user.Username
	}
	if query != "" {
		params["query"] = query
	}
	res, err := Get(url, params, nil)
	if err != nil {
		fmt.Printf("err: %v\n", err)
	}

	var data CmdbHostResponse
	if err := json.Unmarshal(res, &data); err != nil {
		fmt.Printf("数据转化失败:%v", err)
		return nil, err
	} else if data.Code != 0 {
		fmt.Printf("数据获取失败：%v", data.Msg)
		return nil, err
	}
	return data.Data, nil
}

func StorageHostToRedis(hosts []ltwmodels.CmdbHostInfo) error {
	// 缓存主机ip和主机信息的映射关系
	ctx := context.Background()
	for _, h := range hosts {
		user := h.AdminUser
		if h.Rsa == "" {
			t := strings.Split(h.AdminUser, ":")
			if len(t) < 2 {
				continue
			}
			rsa, err := getSecret(t[0], t[1])
			if err != nil {
				return err
			}
			h.Rsa = rsa

			user = strings.Split(t[0], "_")[1]
		}

		m := map[string]string{
			"hostname": h.HostName,
			"port":     h.Port,
			"user":     user,
			"rsa":      h.Rsa,
			"passwd":   h.Passwd,
		}
		err := storage.Redis.HSet(ctx, "host_"+h.Ip, m).Err()
		if err != nil {
			return err
		}
	}
	return nil
}
