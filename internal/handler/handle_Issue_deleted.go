package handler

import (
	"errors"
	"whenchangesth/pkg"
)

// handleIssueDeleted 处理JIRA问题删除事件
func handleIssueDeleted(payload interface{}) error {
	pl, ok := payload.(pkg.IssueDeletedPayload)
	if !ok {
		return errors.New("invalid payload type for issue deleted")
	}

	fields := pl.Issue.Fields
	assigneeNumber := getPhoneNumberWithFallback(fields.Assignee.DisplayName)
	reporterNumber := getPhoneNumberWithFallback(fields.Reporter.DisplayName)

	args := &eventArgs{
		eventType:     EventDelete,
		summaryKeyID:  pl.Issue.Key,
		operator:      pl.User.DisplayName,
		assigneePhone: assigneeNumber,
		reporterPhone: reporterNumber,
		summary:       fields.Summary,
	}

	PushEventArgumentsAndPhones(args)

	return nil
}
