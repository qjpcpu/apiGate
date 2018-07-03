package middlewares

import (
	"github.com/gin-gonic/gin"
	"github.com/qjpcpu/apiGate/global"
	"net/http"
	"strings"
)

func CorsHandle() gin.HandlerFunc {
	devMode := global.G.DevMode
	return func(c *gin.Context) {
		from := c.Request.Header.Get("Origin")
		if from != "" {
			if strings.Contains(from, global.G.Domain) || devMode {
				// ADD CODE HERE: 跨域处理
				c.Writer.Header().Set("Access-Control-Allow-Origin", from)
				c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

				c.Writer.Header().Set("Access-Control-Allow-Headers", "content-type, X-Requested-With")
				c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, DELETE, PUT,PATCH")
				if c.Request.Method == "OPTIONS" {
					c.AbortWithStatus(http.StatusOK)
				}
			}
		}
	}
}
