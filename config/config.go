package config

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

// Config 机器人配置
type Config struct {
	AppID     string `yaml:"appid"`
	AppSecret string `yaml:"secret"`
	Host      string `yaml:"host"`
	Port      int    `yaml:"port"`
	Path      string `yaml:"path"`
	Debug     bool   `yaml:"debug"`
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Host:  "0.0.0.0",
		Port:  9000,
		Path:  "/qqbot",
		Debug: true,
	}
}

// Load 从 config.yaml 加载配置
func Load(path string) *Config {
	cfg := DefaultConfig()

	content, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("读取配置文件失败: %v", err)
	}

	if err = yaml.Unmarshal(content, cfg); err != nil {
		log.Fatalf("解析配置文件失败: %v", err)
	}

	if cfg.AppID == "" || cfg.AppSecret == "" {
		log.Fatalln("appid 和 secret 不能为空，请在 config.yaml 中配置")
	}

	return cfg
}
