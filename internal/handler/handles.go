package handler

import (
	"errors"
	"fmt"
	"github.com/youxihu/dingtalk/dingtalk"
	"log"
	"whenchangesth/internal/dingcfg"
	"whenchangesth/pkg"
)

var (
	cfg   *dingcfg.Config
	err   error
	title = "JIRA事件通知"
)

// 初始化函数，在程序启动时解析配置
func init() {
	// 解析钉钉配置
	cfg, err = dingcfg.ParseConfig()
	if err != nil {
		log.Fatal("DingTalk配置解析失败:", err)
	}
}

func handleIssueCreated(payload interface{}) error {
	pl, ok := payload.(pkg.IssueCreatedPayload)
	if !ok {
		return fmt.Errorf("expected payload type IssueCreatedPayload, got %T", payload)
	}

	// 检查是否有 ChangeLog 并处理
	if pl.ChangeLog != nil && len(pl.ChangeLog.Items) > 0 {
		for _, item := range pl.ChangeLog.Items {
			// 仅处理 "assignee" 字段的变更
			if item.Field != "assignee" {
				continue
			}

			assigneeNumber, err := dingcfg.GetPhoneNumber(pl.Issue.Fields.Assignee.DisplayName)
			if err != nil {
				log.Printf("Error getting phone number for assignee: %v", err)
				assigneeNumber = "" // 如果获取失败，设置为空字符串
			}

			// 获取 reporter 的电话号码
			reporterNumber, err := dingcfg.GetPhoneNumber(pl.Issue.Fields.Reporter.DisplayName)
			if err != nil {
				log.Printf("Error getting phone number for reporter: %v", err)
				reporterNumber = "" // 如果获取失败，设置为空字符串
			}

			// 构造通知内容
			createText := fmt.Sprintf(`
### **事件通知: %s %s创建**
- **摘要名称**: %s
- **经办人**:  → **%s**
- **操作人**: %s
- **JIRA对应摘要地址**: [%s](https://hzbxtx.atlassian.net/browse/%s?linkSource=email)
---
###### 备注:
<small>本通知可作为正式审计记录，由自动化运维系统发送。</small>

###### 提醒:
<small>请相关责任人及时确认。如有异议，请联系运维团队。</small>

@%s
@%s
`,
				pl.Issue.Fields.Type.Name,
				pl.Issue.Key,
				pl.Issue.Fields.Summary,
				item.ToString,
				pl.User.DisplayName,
				pl.Issue.Fields.Summary,
				pl.Issue.Key,
				assigneeNumber,
				reporterNumber)

			// 构造 atMobiles
			atMobiles := []string{}
			if assigneeNumber != "" {
				atMobiles = append(atMobiles, assigneeNumber)
			}
			if reporterNumber != "" {
				atMobiles = append(atMobiles, reporterNumber)
			}

			// 发送通知
			err = dingtalk.SendDingDingNotification(cfg.Token, cfg.Secret, title, createText, atMobiles, false)
			if err != nil {
				fmt.Printf("Error sending notification: %v\n", err)
			}
		}
	}

	return nil
}

// handleIssueDeleted 处理 IssueDeletedEvent 事件
func handleIssueDeleted(payload interface{}) error {
	pl, ok := payload.(pkg.IssueDeletedPayload)
	if !ok {
		return errors.New("invalid payload type for issue deleted")
	}

	// 获取 assignee 的电话号码
	assigneeNumber, err := dingcfg.GetPhoneNumber(pl.Issue.Fields.Assignee.DisplayName)
	if err != nil {
		log.Printf("Error getting phone number for assignee: %v", err)
		assigneeNumber = "" // 如果获取失败，设置为空字符串
	}

	// 获取 reporter 的电话号码
	reporterNumber, err := dingcfg.GetPhoneNumber(pl.Issue.Fields.Reporter.DisplayName)
	if err != nil {
		log.Printf("Error getting phone number for reporter: %v", err)
		reporterNumber = "" // 如果获取失败，设置为空字符串
	}

	// 构造通知内容
	deleteText := fmt.Sprintf(`
### **事件通知: %s %s被删除**
- **摘要名称**: ~~%s~~
- **操作人**: %s
---
###### 备注:
<small>本通知可作为正式审计记录，由自动化运维系统发送。</small>

###### 提醒:
<small>请相关责任人及时确认。如有异议，请联系运维团队。</small>

@%s
@%s
`,
		pl.Issue.Fields.Type.Name,
		pl.Issue.Key,
		pl.Issue.Fields.Summary,
		pl.User.DisplayName,
		assigneeNumber,
		reporterNumber)

	// 构造 atMobiles
	atMobiles := []string{}
	if assigneeNumber != "" {
		atMobiles = append(atMobiles, assigneeNumber)
	}
	if reporterNumber != "" {
		atMobiles = append(atMobiles, reporterNumber)
	}

	// 发送通知
	err = dingtalk.SendDingDingNotification(cfg.Token, cfg.Secret, title, deleteText, atMobiles, false)
	if err != nil {
		fmt.Printf("Error sending notification: %v\n", err)
	}

	return nil
}

// 有关Summary变更的操作器
func handleIssueUpdated(payload interface{}) error {
	// 根据 payload 类型进行处理
	switch payload.(type) {
	case pkg.IssueUpdatedPayload:
		pl, _ := payload.(pkg.IssueUpdatedPayload)
		// 如果有 ChangeLog，记录变更详情
		if pl.ChangeLog != nil && len(pl.ChangeLog.Items) > 0 {
			for _, item := range pl.ChangeLog.Items {
				// 如果 Field 不是 "Status", "Assignee", 或 "Reporter"，则跳过
				if !(item.Field == "reporter") {
					continue
				}
				// 构造通知内容
				// 获取 assignee 的电话号码
				assigneeNumber, err := dingcfg.GetPhoneNumber(pl.Issue.Fields.Assignee.DisplayName)
				if err != nil {
					log.Printf("Error getting phone number for assignee: %v", err)
					assigneeNumber = "" // 如果获取失败，设置为空字符串
				}

				// 获取 reporter 的电话号码
				reporterNumber, err := dingcfg.GetPhoneNumber(pl.Issue.Fields.Reporter.DisplayName)
				if err != nil {
					log.Printf("Error getting phone number for reporter: %v", err)
					reporterNumber = "" // 如果获取失败，设置为空字符串
				}
				reporterText := fmt.Sprintf(`
### **事件通知:  %s 报告人变更**
- **摘要名称**: %s
- **状态**: %s
- **报告人**: ~~%s~~ → **%s**
- **操作人**: %s
- **JIRA对应摘要地址**: [%s](https://hzbxtx.atlassian.net/browse/%s?linkSource=email)
---
###### 备注:
<small>本通知可作为正式审计记录，由自动化运维系统发送。</small>

###### 提醒:
<small>请相关责任人及时确认。如有异议，请联系运维团队。</small>

@%s
@%s
`,
					pl.Issue.Key,
					pl.Issue.Fields.Summary,
					pl.Issue.Fields.Status.Name,
					item.FromString,
					item.ToString,
					pl.User.DisplayName,
					pl.Issue.Fields.Summary,
					pl.Issue.Key,
					assigneeNumber,
					reporterNumber)

				// 构造 atMobiles
				atMobiles := []string{}
				if assigneeNumber != "" {
					atMobiles = append(atMobiles, assigneeNumber)
				}
				if reporterNumber != "" {
					atMobiles = append(atMobiles, reporterNumber)
				}

				// 发送通知
				err = dingtalk.SendDingDingNotification(cfg.Token, cfg.Secret, title, reporterText, atMobiles, false)
				if err != nil {
					fmt.Printf("Error sending notification: %v\n", err)
				}
			}
		}

	case pkg.IssueAssignedPayload:
		pl, _ := payload.(pkg.IssueAssignedPayload)
		// 如果有 ChangeLog，记录变更详情
		if pl.ChangeLog != nil && len(pl.ChangeLog.Items) > 0 {
			for _, item := range pl.ChangeLog.Items {
				// 如果 Field 不是 "Status", "Assignee", 或 "Reporter"，则跳过
				if !(item.Field == "assignee") {
					continue
				}
				assigneeNumber, err := dingcfg.GetPhoneNumber(pl.Issue.Fields.Assignee.DisplayName)
				if err != nil {
					log.Printf("Error getting phone number for assignee: %v", err)
					assigneeNumber = "" // 如果获取失败，设置为空字符串
				}

				// 获取 reporter 的电话号码
				reporterNumber, err := dingcfg.GetPhoneNumber(pl.Issue.Fields.Reporter.DisplayName)
				if err != nil {
					log.Printf("Error getting phone number for reporter: %v", err)
					reporterNumber = "" // 如果获取失败，设置为空字符串
				}

				// 判断 item.FromString 是否为空
				assigneeChange := ""
				if item.FromString == "" {
					assigneeChange = fmt.Sprintf("→ **%s**", item.ToString)
				} else {
					assigneeChange = fmt.Sprintf("~~%s~~ → **%s**", item.FromString, item.ToString)
				}

				// 构造通知内容
				assigneeText := fmt.Sprintf(`
### **事件通知: %s 经办人变更**
- **摘要名称**:  %s
- **状态**: %s
- **经办人**: %s
- **操作人**: %s
- **JIRA对应摘要地址**: [%s](https://hzbxtx.atlassian.net/browse/%s?linkSource=email)
---
###### 备注:
<small>本通知可作为正式审计记录，由自动化运维系统发送。</small>

###### 提醒:
<small>请相关责任人及时确认。如有异议，请联系运维团队。</small>

@%s
@%s
`,
					pl.Issue.Key,
					pl.Issue.Fields.Summary,
					pl.Issue.Fields.Status.Name,
					assigneeChange,
					pl.User.DisplayName,
					pl.Issue.Fields.Summary,
					pl.Issue.Key,
					assigneeNumber, reporterNumber)

				// 构造 atMobiles
				atMobiles := []string{}
				if assigneeNumber != "" {
					atMobiles = append(atMobiles, assigneeNumber)
				}
				if reporterNumber != "" {
					atMobiles = append(atMobiles, reporterNumber)
				}

				// 发送通知
				err = dingtalk.SendDingDingNotification(cfg.Token, cfg.Secret, title, assigneeText, atMobiles, false)
				if err != nil {
					fmt.Printf("Error sending notification: %v\n", err)
				}
			}
		}

	case pkg.IssueGenericPayload:
		pl, _ := payload.(pkg.IssueGenericPayload)
		// 如果有 ChangeLog，记录变更详情
		if pl.ChangeLog != nil && len(pl.ChangeLog.Items) > 0 {
			for _, item := range pl.ChangeLog.Items {
				// 如果 Field 不是 "Status", "Assignee", 或 "Reporter"，则跳过
				if !(item.Field == "status") {
					continue
				}
				assigneeNumber, err := dingcfg.GetPhoneNumber(pl.Issue.Fields.Assignee.DisplayName)
				if err != nil {
					log.Printf("Error getting phone number for assignee: %v", err)
					assigneeNumber = "" // 如果获取失败，设置为空字符串
				}

				// 获取 reporter 的电话号码
				reporterNumber, err := dingcfg.GetPhoneNumber(pl.Issue.Fields.Reporter.DisplayName)
				if err != nil {
					log.Printf("Error getting phone number for reporter: %v", err)
					reporterNumber = "" // 如果获取失败，设置为空字符串
				}
				statusText := fmt.Sprintf(`
### **事件通知: %s 状态变更**
- **摘要名称**:  %s
- **状态**:  ~~%s~~ → **%s**
- **操作人**: %s
- **JIRA对应摘要地址**: [%s](https://hzbxtx.atlassian.net/browse/%s?linkSource=email)
---
###### 备注:
<small>本通知可作为正式审计记录，由自动化运维系统发送。</small>

###### 提醒:
<small>请相关责任人及时确认。如有异议，请联系运维团队。</small>

@%s
@%s
`,
					pl.Issue.Key,
					pl.Issue.Fields.Summary,
					item.FromString,
					item.ToString,
					pl.User.DisplayName,
					pl.Issue.Fields.Summary,
					pl.Issue.Key,
					assigneeNumber, reporterNumber)

				// 构造 atMobiles
				atMobiles := []string{}
				if assigneeNumber != "" {
					atMobiles = append(atMobiles, assigneeNumber)
				}
				if reporterNumber != "" {
					atMobiles = append(atMobiles, reporterNumber)
				}

				// 发送通知
				err = dingtalk.SendDingDingNotification(cfg.Token, cfg.Secret, title, statusText, atMobiles, false)
				if err != nil {
					fmt.Printf("Error sending notification: %v\n", err)
				}
			}
		}
	}
	return nil
}
