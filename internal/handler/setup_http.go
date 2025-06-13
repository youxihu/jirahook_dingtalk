package handler

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"log"
	"whenchangesth/internal/conf"
)

var (
	dingCfg *conf.DingBotStr
	rdsCfg  *conf.RedisConfig
	err     error

	RedisClient *redis.Client
	ctx         = context.Background()
)

func init() {
	// 解析钉钉配置
	dingCfg, err = conf.ParseDingConfig()
	if err != nil {
		log.Fatal("DingTalk配置解析失败:", err)
	}

	// 解析 Redis 配置
	rdsCfg, err = conf.ParseRedisConfig()
	if err != nil {
		log.Fatal("Redis配置解析失败:", err)
	}

	// 初始化 Redis 客户端
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", rdsCfg.Addr, rdsCfg.Port),
		Password: rdsCfg.Password,
		DB:       rdsCfg.DB,
	})

	// 测试连接
	if _, err := RedisClient.Ping(ctx).Result(); err != nil {
		log.Fatalf("Redis 连接失败: %v", err)
	}
	log.Printf("Redis 已连接: %v", RedisClient)
}

func SetupHTTP() {
	// 创建 Gin 引擎
	r := gin.Default()

	// 注册路由
	r.POST("/jira/webhook", JiraWebhookHandler)

	// 启动 HTTP 服务
	r.Run(":4165")
}
