package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/qjpcpu/apiGate/conf"
	ms "github.com/qjpcpu/apiGate/middlewares"
	"github.com/qjpcpu/apiGate/uri"
	syslog "log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
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
	var servers []*http.Server
	for loop := true; loop; loop = false {
		confobj := conf.Get()
		// service https
		if confobj.SSLEnabled() {
			httpsServ := &http.Server{
				Addr:    confobj.ListenAddr,
				Handler: getEngine(),
			}
			servers = append(servers, httpsServ)
			go func(cert, key string) {
				if err := httpsServ.ListenAndServeTLS(cert, key); err != nil && err != http.ErrServerClosed {
					syslog.Fatalf("listen:%s\n", err)
				}
			}(confobj.SSL.CertFile, confobj.SSL.KeyFile)
			if confobj.SSL.EnableHttpPort != "" {
				httpServ := &http.Server{
					Addr:    confobj.SSL.EnableHttpPort,
					Handler: getEngine(),
				}
				servers = append(servers, httpServ)
				go func() {
					if err := httpServ.ListenAndServe(); err != nil && err != http.ErrServerClosed {
						syslog.Fatalf("listen:%s\n", err)
					}
				}()
			}
			break
		}
		// service http
		httpServ := &http.Server{
			Addr:    confobj.ListenAddr,
			Handler: getEngine(),
		}
		servers = append(servers, httpServ)
		go func() {
			if err := httpServ.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				syslog.Fatalf("listen:%s\n", err)
			}
		}()
	}
	if len(servers) == 0 {
		panic("no servers lanched")
	}

	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGABRT, syscall.SIGALRM, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	syslog.Println("shutdown apiGate...")
	wg := new(sync.WaitGroup)
	for i := range servers {
		wg.Add(1)
		go func(srv *http.Server) {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := srv.Shutdown(ctx); err != nil {
				syslog.Printf("apiGate shutdown:%v", err)
			}
		}(servers[i])
	}
	wg.Wait()
	syslog.Println("apiGate exiting.")
}

func getEngine() *gin.Engine {
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
	return ginengine
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
