package handler

import (
	"fmt"
	"whenchangesth/pkg"
)

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

		args := &eventArgs{
			eventType:     EventCreate,
			summaryKeyID:  pl.Issue.Key,
			operator:      pl.User.DisplayName,
			assigneePhone: assigneeNumber,
			reporterPhone: reporterNumber,
			summary:       fields.Summary,
		}

		PushEventArgumentsAndPhones(args)

	}

	return nil
}
