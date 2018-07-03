package mod

import (
	"fmt"
	"github.com/qjpcpu/apiGate/myrouter"
	"strings"
	"sync"
)

var (
	RouterMutex          *sync.Mutex = &sync.Mutex{}
	_routerIndex         int32       = 0
	_black_uri_routers   [2]*myrouter.Router
	_white_uri_routers   [2]*myrouter.Router
	_normal_uri_routers  [2]*myrouter.Router
	_buildin_uri_routers [2]*myrouter.Router
)

const URI_METHOD = "POST"

type APIType int

const (
	UnkownAPI = iota
	NormalAPI
	WhiteAPI
	BlackAPI
	BuildinAPI
)

type FreqCtrl map[string]int64

type API struct {
	Paths    []APIPath `json:"paths" yaml:"paths"`
	FreqCtrl FreqCtrl  `json:"freq" yaml:"freq"`
}

type APIProxy struct {
	Host   string `json:"host" yaml:"host"`
	Prefix string `json:"trim" yaml:"trim"`
}

func (ap *APIProxy) Scheme() string {
	if strings.HasPrefix(ap.Host, "https://") {
		return "https"
	} else {
		return "http"
	}
}

func (ap *APIProxy) HostWithoutScheme() string {
	if ap.Scheme() == "https" {
		return strings.TrimPrefix(ap.Host, "https://")
	} else {
		return strings.TrimPrefix(ap.Host, "http://")
	}
}

func (fc FreqCtrl) GenRouterSetting(routerPath string) myrouter.HostSetting {
	return myrouter.HostSetting{
		RouterPath: routerPath,
		PathRewrite: func(string) string {
			return routerPath
		},
	}
}

func (ap APIProxy) GenRouterSetting(routerPath string) myrouter.HostSetting {
	hs := myrouter.HostSetting{
		Host:       ap.HostWithoutScheme(),
		RouterPath: routerPath,
	}
	if len(ap.Prefix) > 0 {
		hs.PathRewrite = func(raw string) string {
			return strings.TrimPrefix(raw, ap.Prefix)
		}
	} else {
		hs.PathRewrite = func(raw string) string {
			return raw
		}
	}
	return hs
}

type APIPath struct {
	White  []string  `json:"white_list,omitempty" yaml:"white_list,omitempty"`
	Normal []string  `json:"normal_list,omitempty" yaml:"normal_list,omitempty"`
	Black  []string  `json:"black_list,omitempty" yaml:"black_list,omitempty"`
	Proxy  *APIProxy `json:"proxy,omitempty" yaml:"proxy,omitempty"`
}

func FindUri(router *myrouter.Router, path string) (*myrouter.HostSetting, bool) {
	h, _, _ := router.Lookup(URI_METHOD, path)
	if h != nil {
		hs := h()
		return &hs, true
	} else {
		return nil, false
	}
}

func (api API) ValidateAPI() error {
	var errmsg string
	api.doValidate(&errmsg)
	if errmsg != "" {
		return fmt.Errorf("%s", errmsg)
	}
	return nil
}

func (api API) doValidate(errmsg *string) {
	vr := myrouter.New()
	defer func() {
		if r := recover(); r != nil {
			*errmsg = fmt.Sprintf("%v", r)
		}
	}()
LOOP:
	for loop := true; loop; loop = false {
		for _, group := range api.Paths {
			if group.Proxy.Host == "" {
				*errmsg = "no proxy host"
				break LOOP
			}
			for _, p := range group.White {
				vr.HandlerFunc("POST", "/white"+p, group.Proxy.GenRouterSetting(p))
			}
			for _, p := range group.Black {
				vr.HandlerFunc("POST", "/black"+p, group.Proxy.GenRouterSetting(p))
			}
			for _, p := range group.Normal {
				vr.HandlerFunc("POST", "/normal"+p, group.Proxy.GenRouterSetting(p))
			}
		}
	}
}
