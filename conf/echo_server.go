package conf

import (
	"github.com/gin-gonic/gin"
	"io/ioutil"
)

func startEchoServer(port string) {
	echo := gin.Default()
	echo.Any("/*uri", func(c *gin.Context) {
		var payload interface{}
		if c.Request.Body != nil {
			data, _ := ioutil.ReadAll(c.Request.Body)
			defer c.Request.Body.Close()
			payload = string(data)
		}
		qs := make(map[string]string)
		for k := range c.Request.URL.Query() {
			qs[k] = c.Query(k)
		}
		c.JSON(200, gin.H{
			"method":  c.Request.Method,
			"headers": c.Request.Header,
			"path":    c.Request.URL.Path,
			"query":   qs,
			"body":    payload,
		})
	})
	echo.Run(port)
}
