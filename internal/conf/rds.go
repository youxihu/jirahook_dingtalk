package conf

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type RedisConfig struct {
	Addr     string `yaml:"REDIS_ADDR"`
	Port     string `yaml:"REDIS_PORT"`
	Password string `yaml:"REDIS_PASSWORD"`
	DB       int    `yaml:"REDIS_DB"`
}

func ParseRedisConfig() (*RedisConfig, error) {
	//filePath := "/app-acc/configs/redis.yaml"
	filePath := "/home/youxihu/secret/jira_hook/redis.yaml"
	file, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("读取 Redis 配置文件失败: %v", err)
	}

	var config RedisConfig
	err = yaml.Unmarshal(file, &config)
	if err != nil {
		return nil, fmt.Errorf("解析 Redis 配置文件失败: %v", err)
	}

	return &config, nil
}
