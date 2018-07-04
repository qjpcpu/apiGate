package conf

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/qjpcpu/apiGate/uri"
	"github.com/qjpcpu/log"
	"io/ioutil"
	"os"
	"path/filepath"
)

type Configure struct {
	ListenAddr string `json:"listen_addr" yaml:"listen_addr"`

	RedisConfig *CacheConfig `json:"redis_config" yaml:"redis_config"`
	LogDir      string       `json:"log_dir" yaml:"log_dir"`
	LogFile     string       `json:"log_file" yaml:"log_file"`
	LogLevel    string       `json:"log_lvl" yaml:"log_lvl"`

	ConnTimeout    int64 `json:"conn_timeout" yaml:"conn_timeout"`       // in second
	RequestTimeout int64 `json:"request_timeout" yaml:"request_timeout"` // in second

	FreqCtrlDuration int64 `json:"freq_ctrl_duration" yaml:"freq_ctrl_duration"` // 频控窗口,最小30秒

	API                  uri.API `json:"api_list" yaml:"api_list"`
	SessionExpireSeconds int     `json:"session_max_age" yaml:"session_max_age"`

	// 开发模式下允许跨域,正式环境禁止配置
	DevMode bool `json:"dev_mode" yaml:"dev_mode"`

	Domain string `json:"domain" yaml:"domain"` // eg: .baidu.com
}

func (config Configure) String() string {
	return fmt.Sprintf(`监听端口:  %s
开发模式:  %s
日志目录:  %s
跨域允许:  %s
频控窗口:  %s
频控API:
%s
内置API:
%s
外部API列表:
%s
`,
		config.ListenAddr,
		func() string {
			if config.DevMode {
				return "是"
			} else {
				return "否"
			}
		}(),
		filepath.Join(config.LogDir, config.LogFile),
		config.Domain,
		func() string {
			if config.FreqCtrlDuration >= 30 {
				return fmt.Sprintf("%vs", config.FreqCtrlDuration)
			} else {
				return fmt.Sprintf("%vs(无效,必须大于等于30s)", config.FreqCtrlDuration)
			}
		}(),
		func() string {
			if len(config.API.FreqCtrl) == 0 || config.FreqCtrlDuration < 30 {
				return "(无)"
			}
			var str string
			for k, c := range config.API.FreqCtrl {
				str += fmt.Sprintf("%s  每%v秒%d次\n", k, config.FreqCtrlDuration, c)
			}
			return str
		}(),
		func() string {
			if len(uri.Routers()) == 0 {
				return "(无)"
			}
			var str string
			for p := range uri.Routers() {
				str += p + "\n"
			}
			return str
		}(),
		func() string {
			var str string
			for _, api := range config.API.Paths {
				for _, u := range api.Normal {
					str += fmt.Sprintf("* %s  -->  %s\n", u, api.Proxy.Host)
				}
				for _, u := range api.White {
					str += fmt.Sprintf("o %s  -->  %s\n", u, api.Proxy.Host)
				}
				for _, u := range api.Black {
					str += fmt.Sprintf("x %s  -->  %s\n", u, api.Proxy.Host)
				}
			}
			return str
		}(),
	)
}

func (config *Configure) SetDefaults() error {
	if config.SessionExpireSeconds == 0 {
		config.SessionExpireSeconds = 1800
	}

	if config.FreqCtrlDuration == 0 {
		config.FreqCtrlDuration = 30
	}
	if config.ConnTimeout == 0 {
		config.ConnTimeout = 3
	}
	if config.RequestTimeout == 0 {
		config.RequestTimeout = 10
	}
	if config.LogDir == "" {
		config.LogDir = "./apilogs"
	}
	if config.LogFile == "" {
		config.LogFile = "api-gate.log"
	}
	api := config.API
	if len(api.Paths) == 0 {
		return errors.New("配置错误:无路由定义")
	}
	for url, limit := range api.FreqCtrl {
		if limit < 1 {
			return fmt.Errorf("%s频控每%vs不能小于1次", url, config.FreqCtrlDuration)
		}
	}
	for _, g := range api.Paths {
		if len(g.Black) == 0 && len(g.White) == 0 && len(g.Normal) == 0 {
			return errors.New("配置错误:无路由定义")
		}
		if g.Proxy == nil {
			return errors.New("配置错误:无转发配置proxy")
		}
		if g.Proxy.Host == "" {
			return errors.New("配置错误:转发配置无host配置")
		}
	}
	return nil
}

func LoadConfig(config_filename string, config *Configure) error {
	if config_filename == "" {
		return fmt.Errorf("no config file")
	}
	file, err := os.Open(config_filename)
	if err != nil {
		fmt.Printf("load config from file %s error:%v\n", config_filename, err)
		return err
	}
	defer file.Close()
	data, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, config)
	if err != nil {
		return fmt.Errorf("解析配置文件%s失败:%v", config_filename, err)
	}

	if err := config.API.Validate(); err != nil {
		return err
	}
	return nil
}

func InitConfig(config_filename string) {
	confObj := Get()
	err := LoadConfig(config_filename, confObj)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load config failed![Err:%s]\n", err.Error())
		os.Exit(1)
	}
	if err = confObj.SetDefaults(); err != nil {
		fmt.Fprintf(os.Stderr, "parse config failed![Err:%s]\n", err.Error())
		os.Exit(1)
	}
	log.InitLog(log.LogOption{
		LogFile: filepath.Join(confObj.LogDir, confObj.LogFile),
		Level:   log.ParseLogLevel(confObj.LogLevel),
	})
	if confObj.DevMode {
		fmt.Println("===============APIGate运行在开发模式下!!!=================")
	}
}
