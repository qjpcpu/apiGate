package middlewares

import (
	"github.com/gin-gonic/gin"
	"github.com/qjpcpu/apiGate/conf"
	"github.com/qjpcpu/log"
	"io"
	"net"
	"net/http"
	"time"
)

// 转发请求到最终后端服务
func FinalHandler() gin.HandlerFunc {
	// reuse transport object
	commTransport := &http.Transport{
		Dial: func(network, addr string) (net.Conn, error) {
			return net.DialTimeout(network, addr, time.Duration(conf.Get().ConnTimeout)*time.Second)
		},
		ResponseHeaderTimeout: time.Duration(conf.Get().RequestTimeout) * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		DisableKeepAlives:     true,
	}

	return func(c *gin.Context) {
		setting, _ := getProxySetting(c)
		// Copy request
		outreq := new(http.Request)
		*outreq = *(c.Request)
		outreq.URL.Scheme = setting.Scheme
		outreq.URL.Host = setting.Host
		oldPath := c.Request.URL.Path
		newPath := setting.PathRewrite(c.Request.URL.Path)
		outreq.URL.Path = newPath
		outreq.Proto = "HTTP/1.1"
		outreq.ProtoMajor = 1
		outreq.ProtoMinor = 1
		outreq.Close = true
		// for debug
		log.Infof("%s %s => %s user_id:%v", c.Request.Method, oldPath, outreq.URL.String(), getUserId(c))

		reqStart := time.Now()
		resp, err := commTransport.RoundTrip(outreq)
		if err != nil {
			log.Warningf("net error![uri:%s][err:%s]", c.Request.RequestURI, err.Error())
			RenderThenAbort(c, http.StatusGatewayTimeout, makeResponse(ResStateBackendTimeout, nil))
			c.Set("backend_service_error", err)
			return
		}
		defer resp.Body.Close()

		// log real request time
		c.Set("backend_time", time.Now().Sub(reqStart))

		// catch http 401
		if resp.StatusCode == http.StatusUnauthorized {
			log.Infof("catch 401 of %s://%s%s", setting.Scheme, setting.Host, newPath)
			RenderThenAbort(c, http.StatusUnauthorized, makeResponse(ResStateUnauthorized, nil))
			return
		}
		for resp_header_key, _ := range resp.Header {
			c.Writer.Header().Set(resp_header_key, resp.Header.Get(resp_header_key))
		}
		c.Writer.WriteHeader(resp.StatusCode)
		io.Copy(c.Writer, resp.Body)
	}
}
