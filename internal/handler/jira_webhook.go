package handler

import (
	"log"
	"net/http"
	"whenchangesth/pkg"

	"github.com/gin-gonic/gin"
)

// JiraWebhookHandler 处理 /jira/webhook 的 POST 请求
func JiraWebhookHandler(c *gin.Context) {
	// 读取请求Body
	//body, err := c.GetRawData()
	//if err != nil {
	//	log.Printf("Failed to read request body: %v", err)
	//	c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
	//	return
	//}
	//
	//// 打印原始Body（调试用）
	//log.Printf("Raw request body: %s", string(body))
	//
	//// 注意：需要将body写回，因为c.GetRawData()会消费掉Reader
	//c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
	//使用 Parse 方法解析请求体
	result, err := pkg.Parse(c.Request, getAllEvents()...)
	if err != nil {
		log.Printf("Failed to parse webhook: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid webhook request",
		})
		return
	}
	//fmt.Println("result.(type)", reflect.TypeOf(result))
	// 提取事件类型
	var event pkg.Event
	switch v := result.(type) {
	case pkg.Event:
		event = v
	case pkg.TransitionIssueStatusPayload:
		event = pkg.StatusTransitionEvent
	case pkg.IssueCreatedPayload:
		event = pkg.IssueCreatedEvent
	case pkg.IssueDeletedPayload:
		event = pkg.IssueDeletedEvent
	case pkg.IssueUpdatedPayload:
		event = pkg.IssueUpdatedEvent
	case pkg.IssueGenericPayload:
		event = pkg.IssueUpdatedEvent
	case pkg.IssueAssignedPayload:
		event = pkg.IssueUpdatedEvent
	case pkg.IssueWorkLogCreatedPayload:
		event = pkg.IssueWorkLogEvent
	case pkg.IssueWorkLogUpdatedPayload:
		event = pkg.IssueWorkLogEvent
	case pkg.IssueWorkLogDeletedPayload:
		event = pkg.IssueDeletedEvent
	case pkg.IssueMovedPayload:
		event = pkg.IssueDeletedEvent
	case pkg.IssueClosedPayload:
		event = pkg.IssueDeletedEvent
	case pkg.WorkLogCreatedPayload:
		event = pkg.WorkLogCreatedEvent
	case pkg.WorkLogUpdatedPayload:
		event = pkg.WorkLogUpdatedEvent
	case pkg.WorkLogDeletedPayload:
		event = pkg.WorkLogDeletedEvent
	case pkg.CommentCreatedPayload:
		event = pkg.CommentCreatedEvent
	case pkg.CommentUpdatedPayload:
		event = pkg.CommentUpdatedEvent
	case pkg.CommentDeletedPayload:
		event = pkg.CommentDeletedEvent
	case pkg.LinkCreatedPayload:
		event = pkg.LinkCreatedEvent
	case pkg.LinkDeletedPayload:
		event = pkg.LinkDeletedEvent
	case pkg.UserCreatedPayload:
		event = pkg.UserCreatedEvent
	case pkg.UserUpdatedPayload:
		event = pkg.UserUpdatedEvent
	case pkg.UserDeletedPayload:
		event = pkg.UserDeletedEvent
	case pkg.ProjectCreatedPayload:
		event = pkg.ProjectCreatedEvent
	case pkg.ProjectUpdatedPayload:
		event = pkg.ProjectUpdatedEvent
	case pkg.ProjectDeletedPayload:
		event = pkg.ProjectDeletedEvent
	case pkg.ProjectArchivedPayload:
		event = pkg.ProjectArchivedEvent
	case pkg.ProjectRestoredPayload:
		event = pkg.ProjectRestoredEvent
	case pkg.BoardCreatedPayload:
		event = pkg.BoardCreatedEvent
	case pkg.BoardUpdatedPayload:
		event = pkg.BoardUpdatedEvent
	case pkg.BoardDeletedPayload:
		event = pkg.BoardDeletedEvent
	case pkg.BoardConfigurationChangedPayload:
		event = pkg.BoardConfigurationChangedEvent
	case pkg.SprintCreatedPayload:
		event = pkg.SprintCreatedEvent
	case pkg.SprintUpdatedPayload:
		event = pkg.SprintUpdatedEvent
	case pkg.SprintDeletedPayload:
		event = pkg.SprintDeletedEvent
	case pkg.SprintStartedPayload:
		event = pkg.SprintStartedEvent
	case pkg.SprintClosedPayload:
		event = pkg.SprintClosedEvent
	case pkg.VersionCreatedPayload:
		event = pkg.VersionCreatedEvent
	case pkg.VersionUpdatedPayload:
		event = pkg.VersionUpdatedEvent
	case pkg.VersionDeletedPayload:
		event = pkg.VersionDeletedEvent
	case pkg.VersionReleasedPayload:
		event = pkg.VersionReleasedEvent
	case pkg.VersionUnreleasedPayload:
		event = pkg.VersionUnreleasedEvent
	case pkg.OptionTimeTrackingChangedPayload:
		event = pkg.OptionTimeTrackingChangedEvent
	case pkg.OptionIssueLinksChangedPayload:
		event = pkg.OptionIssueLinksChangedEvent
	case pkg.OptionSubTasksChangedPayload:
		event = pkg.OptionSubTasksChangedEvent
	case pkg.OptionAttachmentsChangedPayload:
		event = pkg.OptionAttachmentsChangedEvent
	case pkg.OptionWatchingChangedPayload:
		event = pkg.OptionWatchingChangedEvent
	case pkg.OptionVotingChangedPayload:
		event = pkg.OptionVotingChangedEvent
	case pkg.OptionUnassignedIssuesChangedPayload:
		event = pkg.OptionUnassignedIssuesChangedEvent
	default:
		log.Printf("Unhandled result type: %T", result)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Unsupported hook type",
		})
		return
	}

	// 根据事件类型调用对应的处理器
	handlerFunc, ok := eventHandlers[event]
	if !ok {
		log.Printf("Unsupported event type: %s", event)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Unsupported event type",
		})
		return
	}

	// 调用处理器并处理结果
	if err := handlerFunc(result); err != nil {
		log.Printf("Error processing event %s: %v", event, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to process event",
		})
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"message": "Webhook received and processed successfully",
	})
}
