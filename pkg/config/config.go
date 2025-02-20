package config

import (
	"fmt"
	"os"
)

// Config 存储应用配置
type Config struct {
	DeepseekAPIKey string
}

// LoadConfig 从环境变量加载配置
func LoadConfig() (*Config, error) {
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("DEEPSEEK_API_KEY environment variable is not set")
	}

	return &Config{
		DeepseekAPIKey: apiKey,
	}, nil
}
