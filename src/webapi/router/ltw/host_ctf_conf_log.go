package ltw

import (
	"github.com/didi/nightingale/v5/src/ltwmodels"
	"github.com/gin-gonic/gin"
	"github.com/toolkits/pkg/ginx"
)

func HostCtfConfLogsGets(c *gin.Context) {
	// 接口：获取categraf操作日志
	ip := ginx.QueryStr(c, "ip", "")
	name := ginx.QueryStr(c, "name", "")
	status := ginx.QueryStr(c, "status", "")
	limit := ginx.QueryInt(c, "limit", 1000)

	total, err := ltwmodels.HostCtfLogsTotal(ip, name, status)
	ginx.Dangerous(err)

	res, err := ltwmodels.GetHostCtfLogs(ip, name, status, limit, ginx.Offset(c, limit))
	ginx.Dangerous(err)

	ginx.NewRender(c).Data(gin.H{
		"list":  res,
		"total": total,
	}, nil)
}
