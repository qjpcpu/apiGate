package middlewares

import (
	"github.com/gin-gonic/gin"
	"github.com/qjpcpu/apiGate/embed"
	"github.com/qjpcpu/apiGate/myrouter"
)

// 是否访问的是middleware自有路由
func BuildinFilter() gin.HandlerFunc {
	return func(c *gin.Context) {
		ib, ok := c.Get("IsBuildinUri")
		if !ok {
			return
		}
		if exists, ok := ib.(bool); !ok || !exists {
			return
		}
		ps, _ := c.Get("ProxySetting")
		setting, _ := ps.(*myrouter.HostSetting)
		embed.Handle(c, setting.RouterName)
	}
}
