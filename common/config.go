package common

import (
	"encoding/json"
	"os"
)

// Config 结构体用于存储配置项
type Config struct {
	MaxTokens int `json:"MaxTokens"`
}

// AppConfig 存储全局配置
var AppConfig Config

// LoadConfig 从给定的文件路径加载配置
func LoadConfig(configFile string) error {
	file, err := os.Open(configFile)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&AppConfig)
	if err != nil {
		return err
	}

	return nil
}
