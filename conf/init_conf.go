package conf

import (
	"encoding/json"
	"fmt"
	"github.com/qjpcpu/apiGate/mod"
	"github.com/qjpcpu/log"
	"io/ioutil"
	"os"
	"path/filepath"
)

func LoadConfig(config_filename string, config *mod.Configure) error {
	if config_filename == "" {
		return fmt.Errorf("no config file")
	}
	fmt.Println("load config from file: ", config_filename)
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
	log.InitLog(log.LogOption{LogFile: filepath.Join(confObj.LogDir, confObj.LogFile), Level: log.ParseLogLevel(confObj.LogLevel)})
	d, _ := json.Marshal(confObj)
	if confObj.DevMode {
		fmt.Println("===============APIGate运行在开发模式下!!!=================")
	}
	fmt.Printf("config:\n%v\n", string(d))
}
