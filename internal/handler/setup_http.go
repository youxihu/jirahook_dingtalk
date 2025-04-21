package handler

import (
	"github.com/gin-gonic/gin"
)

func SetupHTTP() {
	// 创建 Gin 引擎
	r := gin.Default()

	// 注册路由
	r.POST("/jira/webhook", JiraWebhookHandler)

	// 启动 HTTP 服务
	r.Run(":4165")
}
