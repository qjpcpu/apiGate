package uri

import (
	"errors"
	"fmt"
	"github.com/qjpcpu/apiGate/myrouter"
	"github.com/qjpcpu/apiGate/rr"
	"sync/atomic"
	"unsafe"
)

type conf struct {
	black_uris   *myrouter.Router
	white_uris   *myrouter.Router
	normal_uris  *myrouter.Router
	buildin_uris *myrouter.Router
	// 频控
	freq_ctrls *myrouter.Router
	// load balance
	services *rr.Services
}

func newconf() *conf {
	return &conf{
		black_uris:   new(myrouter.Router),
		white_uris:   new(myrouter.Router),
		normal_uris:  new(myrouter.Router),
		buildin_uris: new(myrouter.Router),
		freq_ctrls:   new(myrouter.Router),
		services:     &rr.Services{},
	}
}

var g_conf = newconf()

// 进入此函数认为api合法
func InitUri(api API) bool {
	_conf := newconf()
	for url, limit := range api.FreqCtrl {
		p := FREQ_PATH_PREFIX + url
		hs := myrouter.HostSetting{
			RouterPath: p,
			Data:       limit,
			PathRewrite: func(string) string {
				return url
			},
		}
		_conf.freq_ctrls.HandlerFunc(URI_METHOD, p, hs)
	}
	for _, group := range api.Paths {
		_conf.services.AddCluster(group.Proxy.HostWithoutScheme(), group.Proxy.Cluster)
		for _, p := range group.White {
			hs := group.Proxy.GenRouterSetting(p)
			hs.Scheme = group.Proxy.Scheme()
			p = OUTTER_PATH_PREFIX + p
			_conf.white_uris.HandlerFunc(URI_METHOD, p, hs)
		}
		for _, p := range group.Black {
			hs := group.Proxy.GenRouterSetting(p)
			hs.Scheme = group.Proxy.Scheme()
			p = OUTTER_PATH_PREFIX + p
			_conf.black_uris.HandlerFunc(URI_METHOD, p, hs)
		}
		for _, p := range group.Normal {
			hs := group.Proxy.GenRouterSetting(p)
			hs.Scheme = group.Proxy.Scheme()
			p = OUTTER_PATH_PREFIX + p
			_conf.normal_uris.HandlerFunc(URI_METHOD, p, hs)
		}
	}
	_conf.buildin_uris = initBuildinRouter()

	// update config
	oldPtr := (*unsafe.Pointer)(unsafe.Pointer(&g_conf))
	cOld := unsafe.Pointer(g_conf)
	cNew := unsafe.Pointer(_conf)
	return atomic.CompareAndSwapPointer(oldPtr, cOld, cNew)
}

// hot-update
func Update(api API) error {
	if err := api.Validate(); err != nil {
		return err
	}
	if InitUri(api) {
		return nil
	}
	return errors.New("failed update api")
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
	return FindUri(g_conf.black_uris, path)
}

func FindWhiteUri(host, path string) (*myrouter.HostSetting, bool) {
	path = fmt.Sprintf("%s%s", OUTTER_PATH_PREFIX, path)
	return FindUri(g_conf.white_uris, path)
}

func FindNormalUri(host, path string) (*myrouter.HostSetting, bool) {
	path = fmt.Sprintf("%s%s", OUTTER_PATH_PREFIX, path)
	return FindUri(g_conf.normal_uris, path)
}

func FindBuildinUri(host, path string) (*myrouter.HostSetting, bool) {
	path = fmt.Sprintf("%s%s", BUILDIN_PATH_PREFIX, path)
	return FindUri(g_conf.buildin_uris, path)
}

// host is useless, math rule only by path
func FindFreqUri(host, path string) (*myrouter.HostSetting, bool) {
	path = fmt.Sprintf("%s%s", FREQ_PATH_PREFIX, path)
	return FindUri(g_conf.freq_ctrls, path)
}

func GetCluster(name string) (*rr.Cluster, bool) {
	return g_conf.services.GetCluster(name)
}
