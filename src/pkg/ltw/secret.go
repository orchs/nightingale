package ltw

import (
	"encoding/json"
	"fmt"
	"github.com/didi/nightingale/v5/src/webapi/config"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	ssm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/ssm/v20190923"
)

type response struct {
	SecretName   string
	VersionId    string
	SecretBinary string
	SecretString string
	RequestId    string
}

type ssmBody struct {
	Response response
}

func getSecret(sn, v string) (string, error) {
	// 实例化一个认证对象，入参需要传入腾讯云账户secretId，secretKey,此处还需注意密钥对的保密
	// 密钥可前往https://console.cloud.tencent.com/cam/capi网站进行获取
	credential := common.NewCredential(
		config.C.Ltw.TencentSecretId,
		config.C.Ltw.TencentSecretKey,
	)

	// 实例化一个client选项，可选的，没有特殊需求可以跳过
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "ssm.tencentcloudapi.com"

	// 实例化要请求产品的client对象,clientProfile是可选的
	client, _ := ssm.NewClient(credential, "ap-beijing", cpf)

	// 实例化一个请求对象,每个接口都会对应一个request对象
	request := ssm.NewGetSecretValueRequest()

	request.SecretName = common.StringPtr(sn)
	request.VersionId = common.StringPtr(v)

	// 返回的resp是一个GetSecretValueResponse的实例，与请求对象对应
	response, err := client.GetSecretValue(request)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		fmt.Printf("An API error has returned: %s", err)
		return "", err
	}
	if err != nil {
		return "", err
	}

	var obj ssmBody
	json.Unmarshal([]byte(response.ToJsonString()), &obj)
	return obj.Response.SecretString, nil
}
