package handler

import (
	"fmt"
	"log"
	"whenchangesth/internal/conf"

	"github.com/youxihu/dingtalk/dingtalk"
)

const (
	Title = "JIRA事件通知"
)

// getPhoneNumberWithFallback 获取电话号码，如果失败则返回空字符串
func getPhoneNumberWithFallback(displayName string) string {
	phoneNumber, err := conf.GetPhoneNumber(displayName)
	if err != nil {
		log.Printf("Error getting phone number for %s: %v", displayName, err)
		return ""
	}
	return phoneNumber
}

// buildAtMobiles 构建需要@的手机号列表
func buildAtMobiles(assignee, reporter string) []string {
	atMobiles := []string{}
	if assignee != "" {
		atMobiles = append(atMobiles, assignee)
	}
	if reporter != "" {
		atMobiles = append(atMobiles, reporter)
	}
	return atMobiles
}

// sendDingTalkNotification 发送钉钉通知
func sendDingTalkNotification(content string, atMobiles []string) error {
	return dingtalk.SendDingDingNotification(dingCfg.Token, dingCfg.Secret, Title, content, atMobiles, false)
}

// handleError 统一错误处理
func handleError(err error, message string) error {
	if err != nil {
		log.Printf("%s: %v", message, err)
		return fmt.Errorf("%s: %w", message, err)
	}
	return nil
}
