package config

import (
	"fmt"
	"os"
)

// Config 存储应用配置
type Config struct {
	DeepseekAPIKey string
	AutoExecute    bool
	EnableChat    bool
	Debug         bool
}

// LoadConfig 从环境变量加载配置
func LoadConfig() (*Config, error) {
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	autoExecute := os.Getenv("AUTO_EXECUTE")
	enableChat := os.Getenv("ENABLE_CHAT")
	debug := os.Getenv("DEBUG")
	if autoExecute == "" {
		autoExecute = "false"
	}
	if enableChat == "" {
		enableChat = "false"
	}
	if debug == "" {
		debug = "false"
	}
	if apiKey == "" {
		return nil, fmt.Errorf("DEEPSEEK_API_KEY environment variable is not set")
	}

	return &Config{
		DeepseekAPIKey: apiKey,
		AutoExecute:    autoExecute == "true",
		EnableChat:    enableChat == "true",
		Debug:         debug == "true",
	}, nil
}
