package handler

import (
	"errors"
	"fmt"
	"log"

	"github.com/youxihu/dingtalk/dingtalk"
	"whenchangesth/internal/dingcfg"
	"whenchangesth/pkg"
)

const (
	Title = "JIRA事件通知"
)

var (
	cfg *dingcfg.Config
	err error
)

// 初始化函数，在程序启动时解析配置
func init() {
	// 解析钉钉配置
	cfg, err = dingcfg.ParseConfig()
	if err != nil {
		log.Fatal("DingTalk配置解析失败:", err)
	}
}

// getPhoneNumberWithFallback 获取电话号码，如果失败则返回空字符串
func getPhoneNumberWithFallback(displayName string) string {
	phoneNumber, err := dingcfg.GetPhoneNumber(displayName)
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
	return dingtalk.SendDingDingNotification(cfg.Token, cfg.Secret, Title, content, atMobiles, false)
}

// handleError 统一错误处理
func handleError(err error, message string) error {
	if err != nil {
		log.Printf("%s: %v", message, err)
		return fmt.Errorf("%s: %w", message, err)
	}
	return nil
}

// handleIssueCreated 处理JIRA问题创建事件
func handleIssueCreated(payload interface{}) error {
	pl, ok := payload.(pkg.IssueCreatedPayload)
	if !ok {
		return fmt.Errorf("invalid payload type for issue created: %T", payload)
	}

	if len(pl.ChangeLog.Items) == 0 {
		return nil
	}

	for _, item := range pl.ChangeLog.Items {
		if item.Field != "assignee" {
			continue
		}

		fields := pl.Issue.Fields
		assigneeNumber := getPhoneNumberWithFallback(fields.Assignee.DisplayName)
		reporterNumber := getPhoneNumberWithFallback(fields.Reporter.DisplayName)

		content := fmt.Sprintf(`
### **事件通知: %s %s创建**
- **摘要名称**: %s
- **经办人**: → **%s**
- **操作人**: %s
- **JIRA对应摘要地址**: [%s](https://hzbxtx.atlassian.net/browse/%s?linkSource=email)
---
<details>
<summary>展开查看并确认</summary>

###### 备注:
<small>本通知可作为正式审计记录，由自动化运维系统发送。</small>

###### 提醒:
<small>请相关责任人及时确认。如有异议，请联系运维团队。</small>

@%s
@%s
</details>
`,
			fields.Type.Name,
			pl.Issue.Key,
			fields.Summary,
			item.ToString,
			pl.User.DisplayName,
			fields.Summary,
			pl.Issue.Key,
			assigneeNumber,
			reporterNumber)

		err := sendDingTalkNotification(content, buildAtMobiles(assigneeNumber, reporterNumber))
		if err != nil {
			return handleError(err, "发送创建通知失败")
		}
	}

	return nil
}

// handleIssueDeleted 处理JIRA问题删除事件
func handleIssueDeleted(payload interface{}) error {
	pl, ok := payload.(pkg.IssueDeletedPayload)
	if !ok {
		return errors.New("invalid payload type for issue deleted")
	}

	fields := pl.Issue.Fields
	assigneeNumber := getPhoneNumberWithFallback(fields.Assignee.DisplayName)
	reporterNumber := getPhoneNumberWithFallback(fields.Reporter.DisplayName)

	content := fmt.Sprintf(`
### **事件通知: %s %s被删除**
- **摘要名称**: ~~%s~~
- **操作人**: %s
---
<details>
<summary>展开查看并确认
</summary>

###### 备注:
<small>本通知可作为正式审计记录，由自动化运维系统发送。</small>

###### 提醒:
<small>请相关责任人及时确认。如有异议，请联系运维团队。</small>

@%s
@%s
</details>
`,
		fields.Type.Name,
		pl.Issue.Key,
		fields.Summary,
		pl.User.DisplayName,
		assigneeNumber,
		reporterNumber)

	return handleError(
		sendDingTalkNotification(content, buildAtMobiles(assigneeNumber, reporterNumber)),
		"发送删除通知失败",
	)
}

// handleIssueUpdated 处理JIRA问题更新事件
func handleIssueUpdated(payload interface{}) error {
	switch pl := payload.(type) {
	case pkg.IssueUpdatedPayload:
		return handleReporterUpdate(pl)
	case pkg.IssueAssignedPayload:
		return handleAssigneeUpdate(pl)
	case pkg.IssueGenericPayload:
		return handleStatusUpdate(pl)
	default:
		return fmt.Errorf("unknown payload type: %T", payload)
	}
}

// handleReporterUpdate 处理报告人变更
func handleReporterUpdate(pl pkg.IssueUpdatedPayload) error {
	if len(pl.ChangeLog.Items) == 0 {
		return nil
	}

	for _, item := range pl.ChangeLog.Items {
		if item.Field != "reporter" {
			continue
		}

		fields := pl.Issue.Fields
		assigneeNumber := getPhoneNumberWithFallback(fields.Assignee.DisplayName)
		reporterNumber := getPhoneNumberWithFallback(fields.Reporter.DisplayName)

		content := fmt.Sprintf(`
### **事件通知: %s 报告人变更**
- **摘要名称**: %s
- **状态**: %s
- **报告人**: ~~%s~~ → **%s**
- **操作人**: %s
- **JIRA对应摘要地址**: [%s](https://hzbxtx.atlassian.net/browse/%s?linkSource=email)
---
<details>
<summary>展开查看并确认</summary>

###### 备注:
<small>本通知可作为正式审计记录，由自动化运维系统发送。</small>

###### 提醒:
<small>请相关责任人及时确认。如有异议，请联系运维团队。</small>

@%s
@%s
</details>
`,
			pl.Issue.Key,
			fields.Summary,
			fields.Status.Name,
			item.FromString,
			item.ToString,
			pl.User.DisplayName,
			fields.Summary,
			pl.Issue.Key,
			assigneeNumber,
			reporterNumber)

		err := sendDingTalkNotification(content, buildAtMobiles(assigneeNumber, reporterNumber))
		if err != nil {
			return handleError(err, "发送报告人变更通知失败")
		}
	}

	return nil
}

// handleAssigneeUpdate 处理经办人变更
func handleAssigneeUpdate(pl pkg.IssueAssignedPayload) error {
	if len(pl.ChangeLog.Items) == 0 {
		return nil
	}

	for _, item := range pl.ChangeLog.Items {
		if item.Field != "assignee" {
			continue
		}

		fields := pl.Issue.Fields
		assigneeNumber := getPhoneNumberWithFallback(fields.Assignee.DisplayName)
		reporterNumber := getPhoneNumberWithFallback(fields.Reporter.DisplayName)

		assigneeChange := fmt.Sprintf("→ **%s**", item.ToString)
		if item.FromString != "" {
			assigneeChange = fmt.Sprintf("~~%s~~ → **%s**", item.FromString, item.ToString)
		}

		content := fmt.Sprintf(`
### **事件通知: %s 经办人变更**
- **摘要名称**: %s
- **状态**: %s
- **经办人**: %s
- **操作人**: %s
- **JIRA对应摘要地址**: [%s](https://hzbxtx.atlassian.net/browse/%s?linkSource=email)
---
<details>
<summary>展开查看并确认</summary>

###### 备注:
<small>本通知可作为正式审计记录，由自动化运维系统发送。</small>

###### 提醒:
<small>请相关责任人及时确认。如有异议，请联系运维团队。</small>

@%s
@%s
</details>
`,
			pl.Issue.Key,
			fields.Summary,
			fields.Status.Name,
			assigneeChange,
			pl.User.DisplayName,
			fields.Summary,
			pl.Issue.Key,
			assigneeNumber,
			reporterNumber)

		err := sendDingTalkNotification(content, buildAtMobiles(assigneeNumber, reporterNumber))
		if err != nil {
			return handleError(err, "发送经办人变更通知失败")
		}
	}

	return nil
}

// handleStatusUpdate 处理状态变更
func handleStatusUpdate(pl pkg.IssueGenericPayload) error {
	if len(pl.ChangeLog.Items) == 0 {
		return nil
	}

	for _, item := range pl.ChangeLog.Items {
		if item.Field != "status" {
			continue
		}

		fields := pl.Issue.Fields
		assigneeNumber := getPhoneNumberWithFallback(fields.Assignee.DisplayName)
		reporterNumber := getPhoneNumberWithFallback(fields.Reporter.DisplayName)

		content := fmt.Sprintf(`
### **事件通知: %s 状态变更**
- **摘要名称**: %s
- **状态**: ~~%s~~ → **%s**
- **操作人**: %s
- **JIRA对应摘要地址**: [%s](https://hzbxtx.atlassian.net/browse/%s?linkSource=email)
---
<details>
<summary>展开查看并确认</summary>

###### 备注:
<small>本通知可作为正式审计记录，由自动化运维系统发送。</small>

###### 提醒:
<small>请相关责任人及时确认。如有异议，请联系运维团队。</small>

@%s
@%s
</details>
`,
			pl.Issue.Key,
			fields.Summary,
			item.FromString,
			item.ToString,
			pl.User.DisplayName,
			fields.Summary,
			pl.Issue.Key,
			assigneeNumber,
			reporterNumber)

		err := sendDingTalkNotification(content, buildAtMobiles(assigneeNumber, reporterNumber))
		if err != nil {
			return handleError(err, "发送状态变更通知失败")
		}
	}

	return nil
}
