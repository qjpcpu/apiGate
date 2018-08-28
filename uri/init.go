package uri

import (
	"fmt"
	"github.com/qjpcpu/apiGate/myrouter"
	"github.com/qjpcpu/apiGate/rr"
	"github.com/qjpcpu/atomswitch"
)

var (
	_black_uri_routers   = atomswitch.NewAtomicSwitcher(new(myrouter.Router))
	_white_uri_routers   = atomswitch.NewAtomicSwitcher(new(myrouter.Router))
	_normal_uri_routers  = atomswitch.NewAtomicSwitcher(new(myrouter.Router))
	_buildin_uri_routers = atomswitch.NewAtomicSwitcher(new(myrouter.Router))
	// 频控
	freq_ctrl_routers = atomswitch.NewAtomicSwitcher(new(myrouter.Router))
	// load balance
	services = atomswitch.NewAtomicSwitcher(rr.Services{})
)

// 进入此函数认为api合法
func InitUri(api API) {
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
	_buildin_uri_routers.Put(_buildin_uri_router)
	_black_uri_routers.Put(_black_uri_router)
	_white_uri_routers.Put(_white_uri_router)
	_normal_uri_routers.Put(_normal_uri_router)
	freq_ctrl_routers.Put(_freq_uri_router)
	services.Put(_services)
}

// hot-update
func Update(api API) error {
	if err := api.Validate(); err != nil {
		return err
	}
	InitUri(api)
	return nil
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
	path = fmt.Sprintf("%s%s", OUTTER_PATH_PREFIX, path)
	return FindUri(_black_uri_routers.Get().(*myrouter.Router), path)
}

func FindWhiteUri(host, path string) (*myrouter.HostSetting, bool) {
	path = fmt.Sprintf("%s%s", OUTTER_PATH_PREFIX, path)
	return FindUri(_white_uri_routers.Get().(*myrouter.Router), path)
}

func FindNormalUri(host, path string) (*myrouter.HostSetting, bool) {
	path = fmt.Sprintf("%s%s", OUTTER_PATH_PREFIX, path)
	return FindUri(_normal_uri_routers.Get().(*myrouter.Router), path)
}

func FindBuildinUri(host, path string) (*myrouter.HostSetting, bool) {
	path = fmt.Sprintf("%s%s", BUILDIN_PATH_PREFIX, path)
	return FindUri(_buildin_uri_routers.Get().(*myrouter.Router), path)
}

// host is useless, math rule only by path
func FindFreqUri(host, path string) (*myrouter.HostSetting, bool) {
	path = fmt.Sprintf("%s%s", FREQ_PATH_PREFIX, path)
	return FindUri(freq_ctrl_routers.Get().(*myrouter.Router), path)
}

func GetCluster(name string) (*rr.Cluster, bool) {
	return services.Get().(rr.Services).GetCluster(name)
}
