package ltw

import (
	"github.com/didi/nightingale/v5/src/ltwmodels"
	"github.com/gin-gonic/gin"
	"github.com/toolkits/pkg/ginx"
)

func VoiceAdd(c *gin.Context) {
	// 接口：添加语音记录
	var v ltwmodels.Voice
	ginx.BindJSON(c, &v)

	err := v.Add()
	ginx.Dangerous(err)

	ginx.NewRender(c).Message("")
}

func VoiceGet(c *gin.Context) {
	// 接口：查询语音记录
	cid := ginx.UrlParamStr(c, "cid")

	lst, err := ltwmodels.VoiceGetByCallId(cid)
	ginx.Dangerous(err)

	ginx.NewRender(c).Data(lst, nil)
}
