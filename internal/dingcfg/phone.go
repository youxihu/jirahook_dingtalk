package dingcfg

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

// ParsePhone 解析 secret.yaml 文件并返回一个 map[string]string
func ParsePhone() (map[string]string, error) {

	filePath := "/app-acc/phonenumber/secret.yaml"

	// 读取文件内容
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	// 解析 YAML 文件为 map[string]string
	var phoneMap map[string]string
	if err := yaml.Unmarshal(data, &phoneMap); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %v", err)
	}

	return phoneMap, nil
}

// GetPhoneNumber 根据 key 查找对应的电话号码
func GetPhoneNumber(key string) (string, error) {
	// 解析 YAML 文件
	phoneMap, err := ParsePhone()
	if err != nil {
		return "", fmt.Errorf("failed to parse phone config: %v", err)
	}

	// 查找电话号码
	phoneNumber, exists := phoneMap[key]
	if !exists {
		// 如果未找到，返回空字符串
		return "", nil
	}

	return phoneNumber, nil
}
