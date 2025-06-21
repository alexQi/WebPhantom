package config

import (
	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	Sort int // 排序，用于多个配置文件的加载
	Name string
}

type ConfigSetting struct {
	JsonPath string
	YamlPath string
	EnvPath  string
}

func Init(setting *ConfigSetting) {
	// 读取json配置
	if len(setting.JsonPath) > 0 {
		if err := setJsonConfig(setting.JsonPath); err != nil {
			log.Fatalln("初始化配置文件出错", err.Error())
		}
	}
	// 读取yaml
	if len(setting.YamlPath) > 0 {
		if err := setYamlConfig(setting.YamlPath); err != nil {
			log.Fatalln("初始化YAML配置文件出错", err.Error())
		}
	}
	// 读取env
	if len(setting.EnvPath) > 0 {
		if err := setEnvConfig(setting.EnvPath); err != nil {
			log.Fatalln("载入环境变量出错", err.Error())
		}
	}
}

func setJsonConfig(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil
	}
	exsit, _ := pathExists(absPath)
	if exsit == true {
		fileInfoList, err := os.ReadDir(absPath)
		if err != nil {
			return err
		}
		for i := range fileInfoList {
			viper.SetConfigFile(absPath + "/" + fileInfoList[i].Name())
			if err := viper.MergeInConfig(); err != nil {
				return err
			}
			// watchConfig()
		}
	}

	return nil
}

// setYamlConfig 读取config文件夹下的配置
func setYamlConfig(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil
	}
	exsit, _ := pathExists(absPath)
	if exsit == true {
		fileInfoList, err := os.ReadDir(absPath)
		if err != nil {
			return err
		}
		for i := range fileInfoList {
			viper.SetConfigFile(absPath + "/" + fileInfoList[i].Name())
			if err := viper.MergeInConfig(); err != nil {
				return err
			}
			// watchConfig()
		}
	}

	return nil
}

// setEnvConfig
func setEnvConfig(file string) error {
	// 读取系统变量
	viper.AutomaticEnv()
	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)
	// 读取env 文件
	envViper := viper.New()
	// 新建viper用于存储env
	absPath, err := filepath.Abs(file)
	if err != nil {
		return nil
	}
	exsit, _ := pathExists(absPath)
	if exsit == true {
		// 读取.env文件环境变量
		envViper.SetConfigFile(absPath)
		if err := envViper.ReadInConfig(); err != nil {
			return err
		}
	}
	// 配置合并到viper
	envKeys := envViper.AllKeys()
	for i := range envKeys {
		viper.Set(strings.Replace(envKeys[i], "_", ".", 1), envViper.Get(envKeys[i]))
	}

	return nil
}

// 监听配置文件是否改变,用于热更新
func watchConfig() {
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		logrus.Printf("配置文件修改更新: %s\n", e.Name)
	})
}

func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
