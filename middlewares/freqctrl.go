package middlewares

import (
	"github.com/gin-gonic/gin"
	"github.com/qjpcpu/apiGate/global"
	"github.com/qjpcpu/log"
	"net/http"
)

func FreqChecker() gin.HandlerFunc {
	return func(c *gin.Context) {
		userId := c.Request.Header.Get(global.COMM_USER_ID)
		hs, needCtrl := global.FindFreqUri(c.Request.Host, c.Request.URL.Path)
		if !needCtrl || hs == nil || hs.RouterPath == "" {
			log.Debugf("not found freq control for %s", c.Request.URL.Path)
			return
		}
		limit := hs.Data.(int64)
		if limit < 1 {
			return
		}
		if userId == "" {
			return
		}
		if global.G.FreqCtrlDuration < 30 {
			return
		}
		freq := global.GetFreqCtrl(limit, global.G.FreqCtrlDuration)
		if f := freq.Tick(userId, hs.RouterPath); f > 1.0 {
			log.Infof("[frequency exceeded],target:%s path:%s hit:%v ", userId, c.Request.URL.Path, f)
			RenderThenAbort(c, http.StatusTooManyRequests, makeResponse(ResStateReqExceeded, nil))
			return
		}
	}
}
