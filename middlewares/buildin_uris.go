package middlewares

import (
	"github.com/gin-gonic/gin"
	"github.com/qjpcpu/apiGate/myrouter"
	"github.com/qjpcpu/apiGate/uri"
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
		ps, _ := c.Get(gin_context_proxysetting)
		setting, _ := ps.(*myrouter.HostSetting)
		uri.Handle(c, setting.RouterName)
	}
}
