package global

import (
	"encoding/json"
	"fmt"
	"github.com/qjpcpu/apiGate/mod"
	"github.com/qjpcpu/log"
	"io/ioutil"
	"os"
	"path/filepath"
)

func GetConfig(config *mod.Configure) error {
	if G_conf_file == "" {
		return fmt.Errorf("no config file")
	}
	fmt.Println("load config from file: ", G_conf_file)
	file, err := os.Open(G_conf_file)
	if err != nil {
		fmt.Printf("load config from file %s error:%v\n", G_conf_file, err)
		return err
	}
	defer file.Close()
	data, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, config)
	if err != nil {
		return fmt.Errorf("解析配置文件%s失败:%v", G_conf_file, err)
	}

	// 简单校验uri
	if err := validateConfig(config); err != nil {
		return err
	}
	return nil
}

func validateConfig(config *mod.Configure) error {
	if err := config.API.ValidateAPI(); err != nil {
		return err
	}
	return nil
}

func InitConfig() {
	err := GetConfig(&G)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load config failed![Err:%s]\n", err.Error())
		os.Exit(1)
	}
	if err = (&G).AdjustConfiguration(); err != nil {
		fmt.Fprintf(os.Stderr, "parse config failed![Err:%s]\n", err.Error())
		os.Exit(1)
	}
	log.InitLog(log.LogOption{LogFile: filepath.Join(G.LogDir, G.LogFile), Level: log.ParseLogLevel(G.LogLevel)})
	d, _ := json.Marshal(&G)
	if G.DevMode {
		fmt.Println("===============APIGate运行在开发模式下!!!=================")
	}
	fmt.Printf("config:\n%v\n", string(d))
}
