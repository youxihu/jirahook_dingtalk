package handler

import (
	"fmt"
	"time"
	"whenchangesth/internal/objects"
	"whenchangesth/pkg"
)

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

		args := &eventArgs{
			eventType:     EventUpdateReport,
			summaryKeyID:  pl.Issue.Key,
			operator:      pl.User.DisplayName,
			assigneePhone: assigneeNumber,
			reporterPhone: reporterNumber,
			rptFrom:       item.FromString,
			rptTo:         item.ToString,
			status:        "",
			summary:       fields.Summary,
		}

		PushEventArgumentsAndPhones(args)

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

		args := &eventArgs{
			eventType:      EventUpdateAssigner,
			summaryKeyID:   pl.Issue.Key,
			operator:       pl.User.DisplayName,
			assigneePhone:  assigneeNumber,
			reporterPhone:  reporterNumber,
			assignerFromTo: assigneeChange,
			status:         fields.Status.Name,
			statusFrom:     "",
			statusTo:       "",
			summary:        fields.Summary,
		}

		PushEventArgumentsAndPhones(args)

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

		args := &eventArgs{
			eventType:     EventUpdateStatus,
			summaryKeyID:  pl.Issue.Key,
			operator:      pl.User.DisplayName,
			assigneePhone: assigneeNumber,
			reporterPhone: reporterNumber,
			rptFrom:       "",
			rptTo:         "",
			status:        fields.Status.Name,
			statusFrom:    item.FromString,

			statusTo: item.ToString,
			summary:  fields.Summary,
		}

		PushEventArgumentsAndPhones(args)

	}

	return nil
}
func toStdTime(ts objects.Timestamp) time.Time {
	return time.Time(ts)
}
