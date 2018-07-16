package conf

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"strings"
)

func startEchoServer(port string) {
	echo := gin.Default()
	echo.Any("/*uri", func(c *gin.Context) {
		var payload interface{}
		if c.Request.Body != nil {
			data, _ := ioutil.ReadAll(c.Request.Body)
			defer c.Request.Body.Close()
			ct := c.Request.Header.Get("content-type")
			if strings.Contains(ct, "json") {
				payload = make(map[string]interface{})
				json.Unmarshal(data, &payload)
			} else {
				payload = string(data)
			}
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
