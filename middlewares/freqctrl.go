package middlewares

import (
	"github.com/gin-gonic/gin"
	"github.com/qjpcpu/apiGate/conf"
	"github.com/qjpcpu/apiGate/uri"
	"github.com/qjpcpu/log"
	"net/http"
)

func FreqChecker() gin.HandlerFunc {
	return func(c *gin.Context) {
		userId := c.Request.Header.Get(conf.COMM_USER_ID)
		hs, needCtrl := uri.FindFreqUri(c.Request.Host, c.Request.URL.Path)
		if !needCtrl || hs == nil || hs.RouterPath == "" {
			log.Debugf("no freq control for %s", c.Request.URL.Path)
			return
		}
		limit := hs.Data.(int64)
		if limit < 1 {
			return
		}
		if userId == "" {
			userId = c.ClientIP()
		}
		if userId == "" {
			return
		}
		if conf.Get().FreqCtrlDuration < 30 {
			return
		}
		freq := conf.GetFreqCtrl(limit, conf.Get().FreqCtrlDuration)
		if f := freq.TickRule(userId, hs.RouterPath); f > 1.0 {
			log.Infof("[frequency exceeded],target:%s path:%s hit:%v ", userId, c.Request.URL.Path, f)
			RenderThenAbort(c, http.StatusTooManyRequests, makeResponse(ResStateReqExceeded, nil))
			return
		}
	}
}
