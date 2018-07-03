package uri

import (
	"github.com/gin-gonic/gin"
	"github.com/qjpcpu/log"
	"net/http"
)

const (
	BUILDIN_PATH_PREFIX = "/_b"
)

type RouterDef struct {
	// 路由内部名称
	Name string
	// 路由url path
	Path string
	// 路由处理函数
	Handler gin.HandlerFunc
}

var routers map[string]RouterDef = make(map[string]RouterDef)

func Routers() map[string]string {
	simple := make(map[string]string)
	for _, r := range routers {
		simple[r.Name] = r.Path
	}
	return simple
}

func SetRouter(uri string, handler gin.HandlerFunc) {
	routers[uri] = RouterDef{
		Name: uri,
		// 统一添加/_b前缀
		Path:    BUILDIN_PATH_PREFIX + uri,
		Handler: handler,
	}
}

func Handle(c *gin.Context, name string) {
	log.Infof("enter %s handler", c.Request.URL.Path)
	if h, ok := routers[name]; ok && h.Handler != nil {
		h.Handler(c)
		c.Abort()
		log.Infof("leave %s handler", c.Request.URL.Path)
	} else {
		log.Errorf("leave %s handler,shouldn't come here.", c.Request.URL.Path)
		c.AbortWithStatus(http.StatusInternalServerError)
	}
}

func GET(uri string, handler gin.HandlerFunc) {
	SetRouter(uri, handler)
}

func POST(uri string, handler gin.HandlerFunc) {
	SetRouter(uri, handler)
}
