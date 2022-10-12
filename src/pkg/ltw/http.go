package ltw

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

func HttpPost(url, readerStr string) ([]byte, error) {
	reader := strings.NewReader(readerStr)
	resp, err := http.Post(url, "application/x-www-form-urlencoded", reader)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil

}

func HttpGet2(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	fmt.Println(string(body))
	return body, nil
}

func HttpGet(url string, params map[string]string, headers map[string]string) ([]byte, error) {
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
