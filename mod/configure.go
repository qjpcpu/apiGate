package mod

import (
	"errors"
	"fmt"
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

	API                  API `json:"api_list" yaml:"api_list"`
	SessionExpireSeconds int `json:"session_max_age" yaml:"session_max_age"`

	// 开发模式下允许跨域,正式环境禁止配置
	DevMode bool `json:"dev_mode" yaml:"dev_mode"`

	Domain string `json:"domain" yaml:"domain"` // eg: .baidu.com
}

func (config Configure) String() string {
	return fmt.Sprintln(
		fmt.Sprintf("[ListenAddr:%s]", config.ListenAddr),
		fmt.Sprintf("[LogDir:%s][LogFile:%s]", config.LogDir, config.LogFile),
	)
}

func (config *Configure) AdjustConfiguration() error {
	if config.SessionExpireSeconds == 0 {
		config.SessionExpireSeconds = 1800
	}

	if config.FreqCtrlDuration < 30 {
		config.FreqCtrlDuration = 30
	}
	if config.ConnTimeout == 0 {
		config.ConnTimeout = 1
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
