package middlewares

import (
	"github.com/gin-gonic/gin"
	"github.com/qjpcpu/log"
)

func AgentFilter() gin.HandlerFunc {
	return func(c *gin.Context) {
		// ADD CODE HERE: useragent处理
		log.Debugf("parse user_agent[%s]", c.Request.UserAgent())
	}
}
