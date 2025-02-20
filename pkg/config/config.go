package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

// Logger 是全局的日志实例
var Logger = logrus.New()

// Config 存储应用配置
type Config struct {
	DeepseekAPIKey string
	AutoExecute    bool
	EnableChat    bool
	Debug         bool
	LogLevel      string
}

// YAMLConfig 表示配置文件的结构
type YAMLConfig struct {
	Deepseek struct {
		APIKey string `yaml:"api_key"`
	} `yaml:"deepseek"`
	AutoExecute bool   `yaml:"auto_execute"`
	EnableChat bool   `yaml:"enable_chat"`
	LogLevel   string `yaml:"log_level"`
}

// LoadConfig 从配置文件和环境变量加载配置
func LoadConfig() (*Config, error) {
	// 首先从配置文件加载默认值
	var yamlConfig YAMLConfig
	configPath := "./config.yaml"

	// 如果在当前目录找不到配置文件，尝试在上级目录查找
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		configPath = filepath.Join("..", "config.yaml")
	}

	// 读取配置文件
	configData, err := os.ReadFile(configPath)
	if err == nil {
		if err := yaml.Unmarshal(configData, &yamlConfig); err != nil {
			return nil, fmt.Errorf("failed to parse config file: %v", err)
		}
	}

	// 从环境变量读取配置，环境变量优先级高于配置文件
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	autoExecute := os.Getenv("AUTO_EXECUTE")
	enableChat := os.Getenv("ENABLE_CHAT")
	logLevel := os.Getenv("LOG_LEVEL")

	// 如果环境变量未设置，使用配置文件中的值
	if apiKey == "" {
		apiKey = yamlConfig.Deepseek.APIKey
	}
	if autoExecute == "" {
		autoExecute = fmt.Sprintf("%v", yamlConfig.AutoExecute)
	}
	if enableChat == "" {
		enableChat = fmt.Sprintf("%v", yamlConfig.EnableChat)
	}
	if logLevel == "" {
		logLevel = yamlConfig.LogLevel
	}
	if logLevel == "" {
		logLevel = "info" // 默认日志级别
	}

	// 如果 API Key 仍然为空，返回错误
	if apiKey == "" {
		return nil, fmt.Errorf("DEEPSEEK_API_KEY not set in environment variables or config file")
	}

	// 设置日志级别
	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		level = logrus.InfoLevel
	}
	Logger.SetLevel(level)

	// 设置日志格式
	Logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	return &Config{
		DeepseekAPIKey: apiKey,
		AutoExecute:    autoExecute == "true",
		EnableChat:    enableChat == "true",
		LogLevel:      logLevel,
	}, nil
}
