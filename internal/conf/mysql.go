package conf

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

// Config 是整个配置文件的结构
type Config struct {
	MySQL MySQLConfig `yaml:"mysql"`
}

// MySQLConfig 定义 mysql 配置部分
type MySQLConfig struct {
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Database string `yaml:"database"`
}

func ParseMySQLConfig() (*MySQLConfig, error) {
	// filePath := "/app-acc/configs/mysql.yaml" // 生产路径
	filePath := "/home/youxihu/secret/jira_hook/mysql.yaml" // 开发路径

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("读取 mysql 配置文件失败: %v", err)
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("解析 mysql 配置文件失败: %v", err)
	}

	return &config.MySQL, nil
}
