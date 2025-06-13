package conf

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

// DingBOtCfg 定义 YAML 文件的结构
type DingBotStr struct {
	Token  string `yaml:"DINGTALK_ACCESS_TOKEN"`
	Secret string `yaml:"DINGTALK_SECRET"`
}

// ParseDingConfig 从指定路径加载 YAML 配置文件
func ParseDingConfig() (*DingBotStr, error) {
	//filePath := "/app-acc/configs/dingcfg.yaml"
	filePath := "/home/youxihu/secret/jira_hook/test.dingcfg.yaml"
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	var dingCfg DingBotStr
	if err := yaml.Unmarshal(data, &dingCfg); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %v", err)
	}

	return &dingCfg, nil
}
