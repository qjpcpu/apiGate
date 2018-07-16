package middlewares

import (
	"github.com/gin-gonic/gin"
	"github.com/qjpcpu/apiGate/conf"
	"net/http"
	"strings"
)

func CorsHandle() gin.HandlerFunc {
	domain := strings.TrimSuffix(conf.Get().Domain, "/")
	return func(c *gin.Context) {
		from := c.Request.Header.Get("Origin")
		if conf.IsDevMode() || domain == "" || strings.HasSuffix(from, domain) {
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
