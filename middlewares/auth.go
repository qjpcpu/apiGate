package middlewares

import (
	"github.com/gin-gonic/gin"
	"github.com/qjpcpu/apiGate/conf"
	"github.com/qjpcpu/log"
	"net/http"
)

// 用户验证
func SessionFilter() gin.HandlerFunc {
	return func(c *gin.Context) {
		session_id, user_id := "", ""
		var cookie *http.Cookie
		var err error
		if IsWhiteUri(c) {
			// Skip this middleware
			return
		}
		// ADD CODE HERE: session/token验证
		cookie, err = c.Request.Cookie(conf.SESSION_ID)
		if err != nil || cookie.Value == "" {
			log.Info("no cookie or valid token found stop")
			RenderThenAbort(c, http.StatusUnauthorized, makeResponse(ResStateUnauthorized, nil))
			return
		}
		// Ok, user already login in
		session_id = cookie.Value
		user_id, err = FetchUser(session_id)
		if err != nil {
			log.Infof("no cookie or valid token found:%v", err)
			RenderThenAbort(c, http.StatusUnauthorized, makeResponse(ResStateUnauthorized, nil))
			return
		}

		writeUserServerSessionInfo(c, user_id, session_id)
		log.Infof("user:%v session:%v passed", user_id, session_id)
	}
}
