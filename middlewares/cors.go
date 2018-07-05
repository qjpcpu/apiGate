package middlewares

import (
	"github.com/gin-gonic/gin"
	"github.com/qjpcpu/apiGate/conf"
	"net/http"
	"strings"
)

func CorsHandle() gin.HandlerFunc {
	devMode := conf.Get().DevMode
	return func(c *gin.Context) {
		from := c.Request.Header.Get("Origin")
		if devMode || strings.Contains(from, conf.Get().Domain) {
			c.Writer.Header().Set("Access-Control-Allow-Origin", from)
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

			c.Writer.Header().Set("Access-Control-Allow-Headers", c.Request.Header.Get("Access-Control-Request-Headers"))
			c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, DELETE, PUT,PATCH")
		}
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
		}
	}
}
