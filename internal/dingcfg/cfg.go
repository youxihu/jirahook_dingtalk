package dingcfg

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

// Config 定义 YAML 文件的结构
type Config struct {
	Token  string `yaml:"DINGTALK_ACCESS_TOKEN"`
	Secret string `yaml:"DINGTALK_SECRET"`
}

// LoadConfig 从指定路径加载 YAML 配置文件
func ParseConfig() (*Config, error) {
	filePath := "/app-acc/dingtalk/secret.yaml"

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %v", err)
	}

	return &config, nil
}
