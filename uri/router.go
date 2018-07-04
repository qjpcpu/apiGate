package uri

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

const (
	BUILDIN_PATH_PREFIX = "/_b"
	OUTTER_PATH_PREFIX  = "/_o"
	FREQ_PATH_PREFIX    = "/_f"
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
	if h, ok := routers[name]; ok && h.Handler != nil {
		h.Handler(c)
		c.Abort()
	} else {
		c.AbortWithStatus(http.StatusInternalServerError)
	}
}

func GET(uri string, handler gin.HandlerFunc) {
	SetRouter(uri, handler)
}

func POST(uri string, handler gin.HandlerFunc) {
	SetRouter(uri, handler)
}
