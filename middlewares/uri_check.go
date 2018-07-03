package middlewares

import (
	"github.com/gin-gonic/gin"
	"github.com/qjpcpu/apiGate/uri"
	"github.com/qjpcpu/log"
	"net/http"
)

// 依次检查是否是黑名单、系统路由、白名单、普通路由
// 黑名单: 终止
// 系统路由: 系统截获处理
// 白名单和普通路由: 转发请求
func UriCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, is_black_uri := uri.FindBlackUri(c.Request.Host, c.Request.URL.Path)
		if is_black_uri {
			log.Debugf("is_black_uri:%v", c.Request.URL.Path)
			RenderThenAbort(c, http.StatusForbidden, makeResponse(ResStateForbidden, nil))
			return
		}
		hs, is_buildin_uri := uri.FindBuildinUri(c.Request.Host, c.Request.URL.Path)
		if is_buildin_uri {
			log.Debugf("is_buildin_uri:%v", c.Request.URL.Path)
			c.Set("IsBuildinUri", is_buildin_uri)
			c.Set("ProxySetting", hs)
			return
		}
		hs, is_white_uri := uri.FindWhiteUri(c.Request.Host, c.Request.URL.Path)
		if is_white_uri {
			c.Set("IsWhiteUri", is_white_uri)
			if hs.Host == "" {
				hs.Host = c.Request.Host
			}
			if hs.Scheme == "" {
				hs.Scheme = "http"
			}
			c.Set("ProxySetting", hs)
			log.Debugf("is_white_uri:%v", c.Request.URL.Path)
			return
		}
		hs, is_normal_uri := uri.FindNormalUri(c.Request.Host, c.Request.URL.Path)
		if is_normal_uri {
			if hs.Host == "" {
				hs.Host = c.Request.Host
			}
			if hs.Scheme == "" {
				hs.Scheme = "http"
			}
			c.Set("ProxySetting", hs)
			log.Debugf("is_normal_uri:%v", c.Request.URL.Path)
			return
		}

		log.Noticef("unkown request: %s%s, halt.", c.Request.Host, c.Request.URL.Path)
		RenderThenAbort(c, http.StatusNotFound, makeResponse(ResStateNotFound, nil))
		return
	}
}
