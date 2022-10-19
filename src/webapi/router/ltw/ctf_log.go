package ltw

import (
	"github.com/didi/nightingale/v5/src/ltwmodels"
	"github.com/gin-gonic/gin"
	"github.com/toolkits/pkg/ginx"
)

func CtfLogsGets(c *gin.Context) {
	query := ginx.QueryStr(c, "query", "")
	status := ginx.QueryStr(c, "status", "")
	action := ginx.QueryStr(c, "action", "")
	limit := ginx.QueryInt(c, "limit", 1000)

	total, err := ltwmodels.HostCtfLogsTotal(query, status, action)
	ginx.Dangerous(err)

	res, err := ltwmodels.GetCtfLogs(query, status, action, limit, ginx.Offset(c, limit))
	ginx.Dangerous(err)

	ginx.NewRender(c).Data(gin.H{
		"list":  res,
		"total": total,
	}, nil)
}
