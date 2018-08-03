package uri

import (
	"fmt"
	"github.com/qjpcpu/apiGate/myrouter"
	"github.com/qjpcpu/apiGate/rr"
	"os"
	"sync"
	"sync/atomic"
)

var (
	RouterMutex          *sync.Mutex = &sync.Mutex{}
	_routerIndex         int32       = 0
	_black_uri_routers   [2]*myrouter.Router
	_white_uri_routers   [2]*myrouter.Router
	_normal_uri_routers  [2]*myrouter.Router
	_buildin_uri_routers [2]*myrouter.Router
	// 频控
	freq_ctrl_routers [2]*myrouter.Router
	// load balance
	services [2]rr.Services
)

func calcRouterIndex() int32 {
	return _routerIndex % 2
}

func InitUri(api API) {
	rindex := (_routerIndex + 1) % 2
	_black_uri_router := myrouter.New()
	_white_uri_router := myrouter.New()
	_normal_uri_router := myrouter.New()
	_freq_uri_router := myrouter.New()
	_services := rr.NewServices()
	for url, limit := range api.FreqCtrl {
		p := FREQ_PATH_PREFIX + url
		hs := myrouter.HostSetting{
			RouterPath: p,
			Data:       limit,
			PathRewrite: func(string) string {
				return url
			},
		}
		_freq_uri_router.HandlerFunc(URI_METHOD, p, hs)
	}
	for _, group := range api.Paths {
		if len(group.Proxy.Host) == 0 {
			fmt.Println("Please set  proxy host")
			os.Exit(1)
		}
		_services.AddCluster(group.Proxy.HostWithoutScheme(), group.Proxy.Cluster)
		for _, p := range group.White {
			hs := group.Proxy.GenRouterSetting(p)
			hs.Scheme = group.Proxy.Scheme()
			p = OUTTER_PATH_PREFIX + p
			_white_uri_router.HandlerFunc(URI_METHOD, p, hs)
		}
		for _, p := range group.Black {
			hs := group.Proxy.GenRouterSetting(p)
			hs.Scheme = group.Proxy.Scheme()
			p = OUTTER_PATH_PREFIX + p
			_black_uri_router.HandlerFunc(URI_METHOD, p, hs)
		}
		for _, p := range group.Normal {
			hs := group.Proxy.GenRouterSetting(p)
			hs.Scheme = group.Proxy.Scheme()
			p = OUTTER_PATH_PREFIX + p
			_normal_uri_router.HandlerFunc(URI_METHOD, p, hs)
		}
	}
	_buildin_uri_router := initBuildinRouter()
	_buildin_uri_routers[rindex] = _buildin_uri_router
	_black_uri_routers[rindex] = _black_uri_router
	_white_uri_routers[rindex] = _white_uri_router
	_normal_uri_routers[rindex] = _normal_uri_router
	freq_ctrl_routers[rindex] = _freq_uri_router
	services[rindex] = _services
	atomic.AddInt32(&_routerIndex, 1)
}

func initBuildinRouter() *myrouter.Router {
	_buildin_uri_router := myrouter.New()
	for name, path_info := range Routers() {
		_buildin_uri_router.HandlerFunc(URI_METHOD, path_info, myrouter.HostSetting{
			RouterPath:  path_info,
			RouterName:  name,
			PathRewrite: myrouter.DefaultPathRewrite,
			Scheme:      "http",
		})
	}
	return _buildin_uri_router
}

func FindBlackUri(host, path string) (*myrouter.HostSetting, bool) {
	rindex := calcRouterIndex()
	path = fmt.Sprintf("%s%s", OUTTER_PATH_PREFIX, path)
	return FindUri(_black_uri_routers[rindex], path)
}

func FindWhiteUri(host, path string) (*myrouter.HostSetting, bool) {
	rindex := calcRouterIndex()
	path = fmt.Sprintf("%s%s", OUTTER_PATH_PREFIX, path)
	return FindUri(_white_uri_routers[rindex], path)
}

func FindNormalUri(host, path string) (*myrouter.HostSetting, bool) {
	rindex := calcRouterIndex()
	path = fmt.Sprintf("%s%s", OUTTER_PATH_PREFIX, path)
	return FindUri(_normal_uri_routers[rindex], path)
}

func FindBuildinUri(host, path string) (*myrouter.HostSetting, bool) {
	rindex := calcRouterIndex()
	path = fmt.Sprintf("%s%s", BUILDIN_PATH_PREFIX, path)
	return FindUri(_buildin_uri_routers[rindex], path)
}

// host is useless, math rule only by path
func FindFreqUri(host, path string) (*myrouter.HostSetting, bool) {
	rindex := calcRouterIndex()
	path = fmt.Sprintf("%s%s", FREQ_PATH_PREFIX, path)
	return FindUri(freq_ctrl_routers[rindex], path)
}

func GetCluster(name string) (*rr.Cluster, bool) {
	rindex := calcRouterIndex()
	return services[rindex].GetCluster(name)
}
