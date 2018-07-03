package middlewares

import (
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/qjpcpu/apiGate/global"
	"github.com/qjpcpu/apiGate/myrouter"
	"github.com/qjpcpu/apiGate/uid"
	"github.com/qjpcpu/log"
	"io/ioutil"
	"net/http"
)

// 从gin context读取user id
func getUserId(c *gin.Context) string {
	iuid, ok := c.Get("UserId")
	if !ok {
		return ""
	}
	userId, ok := iuid.(string)
	if !ok {
		return ""
	}
	return userId
}

// 注意使用时proxysetting必须已被设置
func getProxySetting(c *gin.Context) (*myrouter.HostSetting, error) {
	ps, ok := c.Get("ProxySetting")
	if !ok {
		return nil, errors.New("host setting not found")
	}
	setting, ok := ps.(*myrouter.HostSetting)
	if !ok || setting == nil {
		return nil, errors.New("host setting not found")
	}
	return setting, nil
}

func writeSession(c *gin.Context, value string) {
	cookie := &http.Cookie{
		Name:   global.SESSION_ID,
		Value:  value,
		Path:   "/",
		MaxAge: global.G.SessionExpireSeconds,
		Domain: global.G.Domain,
	}
	http.SetCookie(c.Writer, cookie)
}

func expireSession(c *gin.Context) {
	cookie := &http.Cookie{
		Name:   global.SESSION_ID,
		Value:  "",
		Path:   "/",
		MaxAge: -1, // delete cookie now
		Domain: global.G.Domain,
	}
	http.SetCookie(c.Writer, cookie)
}

func RenderThenAbort(c *gin.Context, code int, obj interface{}) {
	if code == http.StatusUnauthorized {
		expireSession(c)
	}
	c.JSON(code, obj)
	c.Abort()
}

func RedirectThenAbort(c *gin.Context, redirect_to string) {
	c.Redirect(http.StatusFound, redirect_to)
	c.Abort()
}

func BinarySearch(list []string, key string) int {
	var start int = 0
	var end int = len(list) - 1
	var mid int
	for start <= end {
		mid = start + (end-start)/2
		if list[mid] < key {
			start = mid + 1
		} else if list[mid] > key {
			end = mid - 1
		} else {
			return mid
		}
	}
	return -1
}

func IsWhiteUri(c *gin.Context) bool {
	if iwu, exist := c.Get("IsWhiteUri"); exist {
		ex, ok := iwu.(bool)
		return ok && ex
	}
	return false
}

// 写入用户session信息:http response Set-Cookie,服务端session,gin Context UserId,http proxy Header
// 如果未提供指定sessionID则自动生成
func writeUserServerSessionInfo(c *gin.Context, user_id string, session_ids ...string) (string, error) {
	var session_id string
	if len(session_ids) > 0 && session_ids[0] != "" {
		session_id = session_ids[0]
	} else {
		session_id = uid.GenUniqueId()
	}
	c.Request.Header.Set(global.COMM_USER_ID, user_id)
	// extend session
	if err := global.CacheUser(session_id, user_id); err != nil {
		return session_id, err
	}
	writeSession(c, session_id)
	c.Set("UserId", user_id)
	return session_id, nil
}

func ReadRequestJson(c *gin.Context, obj interface{}) error {
	var body []byte
	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		log.Infof("read body err:%s", err.Error())
		return err
	}
	log.Debugf("input param:%s", string(body))

	if err := json.Unmarshal(body, obj); err != nil {
		log.Infof("Unmarshal body error: %s", err)
		return err
	}

	return nil
}
