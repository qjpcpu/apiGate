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

type SSL struct {
	CertFile string `json:"cert,omitempty" yaml:"cert"`
	KeyFile  string `json:"key,omitempty" yaml:"key"`
}

type Configure struct {
	ListenAddr string `json:"listen_addr,omitempty" yaml:"listen_addr"`
	SSL        *SSL   `json:"ssl,omitempty" yaml:"ssl"`

	RedisConfig *CacheConfig `json:"redis_config,omitempty" yaml:"redis_config"`
	LogDir      string       `json:"log_dir,omitempty" yaml:"log_dir"`
	LogFile     string       `json:"log_file,omitempty" yaml:"log_file"`

	ConnTimeout    int64 `json:"conn_timeout,omitempty" yaml:"conn_timeout"`       // in second
	RequestTimeout int64 `json:"request_timeout,omitempty" yaml:"request_timeout"` // in second

	FreqCtrlDuration int64 `json:"freq_ctrl_duration,omitempty" yaml:"freq_ctrl_duration"` // 频控窗口,最小30秒

	API                  uri.API `json:"api_list" yaml:"api_list"`
	SessionExpireSeconds int     `json:"session_max_age,omitempty" yaml:"session_max_age"`

	Domain string `json:"domain,omitempty" yaml:"domain"` // eg: .baidu.com
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
			if IsDevMode() {
				return "是"
			} else {
				return "否"
			}
		}(),
		filepath.Join(config.LogDir, config.LogFile),
		func() string {
			if IsDevMode() || config.Domain == "" {
				return "*"
			} else {
				return config.Domain
			}
		}(),
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

func (config *Configure) SSLEnabled() bool {
	return config.SSL != nil && config.SSL.CertFile != "" && config.SSL.KeyFile != ""
}

func (config *Configure) SetDefaults() error {
	if config.ListenAddr == "" {
		config.ListenAddr = ":8080"
	}
	if config.RedisConfig == nil {
		config.RedisConfig = &CacheConfig{}
	}
	config.RedisConfig.setDefault()

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
		config.LogDir = "./log"
	}
	if config.LogFile == "" {
		config.LogFile = "api.log"
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
	var err error
	confObj := Get()
	if config_filename != "" {
		err = LoadConfig(config_filename, confObj)
		if err != nil {
			fmt.Fprintf(os.Stderr, "load config failed![Err:%s]\n", err.Error())
			os.Exit(1)
		}
	} else {
		// use simplest config
		*confObj = Configure{
			ListenAddr: ":8080",
			RedisConfig: &CacheConfig{
				RedisAddress: "localhost:6379",
			},
			LogDir: "./log",
			API: uri.API{Paths: []uri.APIPath{
				{
					White: []string{"/*any"},
					Proxy: &uri.APIProxy{
						Host: "localhost:6000",
					},
				},
			}},
		}
		data, _ := json.MarshalIndent(confObj, "", "    ")
		fn := "./simplest.conf"
		if _, err = os.Stat(fn); os.IsNotExist(err) {
			ioutil.WriteFile(fn, data, 0644)
			fmt.Printf("no config file found by flag [-c], use simplest config %s:\n%s\n", fn, string(data))
		} else {
			fmt.Printf("no config file found by flag [-c], use simplest config:\n%s\n", string(data))
		}
	}
	if err = confObj.SetDefaults(); err != nil {
		fmt.Fprintf(os.Stderr, "parse config failed![Err:%s]\n", err.Error())
		os.Exit(1)
	}
	var lvl log.Level
	if IsDevMode() {
		lvl = log.DEBUG
	} else {
		lvl = log.INFO
	}
	log.InitLog(log.LogOption{
		LogFile: filepath.Join(confObj.LogDir, confObj.LogFile),
		Level:   lvl,
	})
	if IsDevMode() {
		fmt.Println("===============APIGate运行在开发模式下!=================")
	}
}
