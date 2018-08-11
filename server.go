package main

import (
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/qjpcpu/apiGate/conf"
	ms "github.com/qjpcpu/apiGate/middlewares"
	"github.com/qjpcpu/apiGate/uri"
	"net/http"
	"os"
	"strings"
)

// version information
var (
	g_Version   string
	g_BuildDate string
	g_CommitID  string
)

var (
	g_config_file string
)

func startServer() {
	ginengine := gin.Default()
	// allow *.com CORS
	ginengine.Use(ms.CorsHandle())
	// PrefixFilter: set IsXXUri, ProxySetting, but black uri would be stopped here
	ginengine.Use(ms.UriCheck())
	// PrefixFilter: set agent, no dependency
	ginengine.Use(ms.AgentFilter())
	// PrefixFilter: buildin handlers, depend on UriCheck
	ginengine.Use(ms.BuildinFilter())
	// PrefixFilter: session check, depend on UriCheck,AgentFilter
	ginengine.Use(ms.SessionFilter())
	// PrefixFilter: no dependency
	ginengine.Use(ms.FreqChecker())

	ginengine.GET("/*uri", ms.FinalHandler())
	ginengine.POST("/*uri", ms.FinalHandler())
	ginengine.DELETE("/*uri", ms.FinalHandler())
	ginengine.PUT("/*uri", ms.FinalHandler())
	ginengine.HEAD("/*uri", ms.FinalHandler())
	ginengine.PATCH("/*uri", ms.FinalHandler())
	ginengine.OPTIONS("/*uri", ms.FinalHandler())

	var err error
	if confobj := conf.Get(); confobj.SSLEnabled() {
		if confobj.SSL.RedirectHttpPort != "" {
			go redirectHttp2Https(*confobj.SSL)
		}
		err = ginengine.RunTLS(confobj.ListenAddr, confobj.SSL.CertFile, confobj.SSL.KeyFile)
	} else {
		err = ginengine.Run(confobj.ListenAddr)
	}
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func redirectHttp2Https(sslconfig conf.SSL) {
	if sslconfig.RedirectHttpPort == "" {
		return
	}
	sslport := conf.Get().ListenAddr
	if sslport == ":443" {
		sslport = ""
	}
	httpRouter := gin.Default()
	httpRouter.Any("/*uri", func(c *gin.Context) {
		rurl := c.Request.URL
		rurl.Scheme = "https"
		host := c.Request.Host
		if strings.HasSuffix(host, sslconfig.RedirectHttpPort) {
			host = strings.TrimSuffix(host, sslconfig.RedirectHttpPort)
		}
		rurl.Host = host + sslport
		c.Redirect(http.StatusMovedPermanently, rurl.String())
	})
	if err := httpRouter.Run(sslconfig.RedirectHttpPort); err != nil {
		fmt.Printf("redirect http to https fail:%v\n", err)
	}
}

func parseArgs() {
	var dev_mode bool
	flag.StringVar(&g_config_file, "c", "", "-c config file name")
	flag.BoolVar(&dev_mode, "dev", false, "dev mode")
	var showVersion bool
	flag.BoolVar(&showVersion, "v", false, "-v")
	flag.Parse()
	conf.SetMode(dev_mode)
	if showVersion {
		fmt.Printf("版本号: %s\n编译时间: %s\nCommitID: %s\n", g_Version, g_BuildDate, g_CommitID)
		os.Exit(0)
	}
}

func main() {
	parseArgs()
	if !conf.IsDevMode() {
		gin.SetMode(gin.ReleaseMode)
	}
	conf.InitConfig(g_config_file)
	conf.InitCache()
	conf.InitIDGenerator(conf.Cache())
	uri.InitUri(conf.Get().API)
	fmt.Print(conf.Get().String())
	startServer()
}
