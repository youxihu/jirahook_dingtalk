package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"sync"
	"time"
	"whenchangesth/internal/conf"
)

// 用户ID -> 定时器（用于去重和刷新）
var debounceTimers = make(map[string]*time.Timer)
var timerLock sync.Mutex // 保护 map 并发访问

// Redis 存活时间和定时器延迟时间
const (
	redisTTL   = 3 * time.Minute // Redis 数据存活时间稍长于定时器
	timerDelay = 2 * time.Minute // 定时器延迟执行时间
)
const (
	EventCreate         = "created"
	EventDelete         = "deleted"
	EventUpdateReport   = "updated_report"
	EventUpdateAssigner = "updated_assigner"
	EventUpdateStatus   = "updated_status"
)

type eventArgs struct {
	eventType,
	summaryKeyID,
	operator,
	assigneePhone,
	reporterPhone,
	rptFrom,
	rptTo,
	assignerFromTo,
	status,
	statusFrom,
	statusTo,
	summary string
}

func PushEventArgumentsAndPhones(args *eventArgs) {
	// 构造 Redis Key
	summaryKey := fmt.Sprintf("issue_event_summary:%s:%s", args.eventType, args.operator)
	phoneKey := fmt.Sprintf("issue_event_phone:%s:%s", args.eventType, args.operator)

	// 将 eventArgs 转换为 JSON 存入 Redis List
	eventData, _ := json.Marshal(map[string]string{
		"summaryKeyID":   args.summaryKeyID,
		"summary":        args.summary,
		"rptFrom":        args.rptFrom,
		"rptTo":          args.rptTo,
		"assignerFromTo": args.assignerFromTo,
		"status":         args.status,
		"statusFrom":     args.statusFrom,
		"statusTo":       args.statusTo,
	})
	err := RedisClient.RPush(ctx, summaryKey, eventData).Err()
	if err != nil {
		fmt.Printf("⚠️ 写入 Redis issue_%s 失败: %v\n", args.eventType, err)
	} else {
		RedisClient.Expire(ctx, summaryKey, redisTTL)
	}
	go func() {
		mysqlConf, err := conf.ParseMySQLConfig()
		if err != nil {
			fmt.Printf("❌ 加载 MySQL 配置失败: %v\n", err)
			return
		}

		err = WithMySQL(mysqlConf, func(db *sql.DB) error {
			query := `
            INSERT INTO jirahook_eventdata 
            (event_type, summary_key_id, operator, assignee_phone, reporter_phone, 
            rpt_from, rpt_to, assigner_from_to, status, status_from, status_to, summary) 
            VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

			_, err := db.Exec(query,
				args.eventType,
				args.summaryKeyID,
				args.operator,
				args.assigneePhone,
				args.reporterPhone,
				args.rptFrom,
				args.rptTo,
				args.assignerFromTo,
				args.status,
				args.statusFrom,
				args.statusTo,
				args.summary,
			)
			return err
		})

		if err != nil {
			fmt.Printf("❌ 写入 MySQL jirahook_eventdata 失败: %v\n", err)
		}
	}()

	// 收集手机号并存入 Set
	var phones []string
	if args.assigneePhone != "" {
		phones = append(phones, args.assigneePhone)
	}
	if args.reporterPhone != "" {
		phones = append(phones, args.reporterPhone)
	}
	if len(phones) > 0 {
		interfacePhones := make([]interface{}, len(phones))
		for i, p := range phones {
			interfacePhones[i] = p
		}
		if err := RedisClient.SAdd(ctx, phoneKey, interfacePhones...).Err(); err != nil {
			fmt.Printf("⚠️ 写入 Redis issue_%s_phone 失败: %v\n", args.eventType, err)
		} else {
			RedisClient.Expire(ctx, phoneKey, redisTTL)
		}
	}

	setIssueDebounceTimer(args.eventType, args.operator)
}

func setIssueDebounceTimer(eventType, operator string) {
	timerLock.Lock()
	defer timerLock.Unlock()

	key := fmt.Sprintf("timer_key:%s:%s", eventType, operator)

	if oldTimer, exists := debounceTimers[key]; exists {
		oldTimer.Stop()
	}

	newTimer := time.AfterFunc(timerDelay, func() {
		sendEventSummariesAndNotifications(eventType, operator)
	})

	debounceTimers[key] = newTimer
}

func sendEventSummariesAndNotifications(eventType, operator string) {
	summaryKey := fmt.Sprintf("issue_event_summary:%s:%s", eventType, operator)
	phoneKey := fmt.Sprintf("issue_event_phone:%s:%s", eventType, operator)

	// 获取所有事件数据
	rawEvents, err := RedisClient.LRange(ctx, summaryKey, 0, -1).Result()
	if err != nil || len(rawEvents) == 0 {
		fmt.Printf("❌ 没有找到事件数据: %v\n", err)
		return
	}

	// 解析事件
	var allEvents []map[string]string
	for _, raw := range rawEvents {
		var eventMap map[string]string
		_ = json.Unmarshal([]byte(raw), &eventMap)
		allEvents = append(allEvents, eventMap)
	}

	// 获取手机号
	phones, err := RedisClient.SMembers(ctx, phoneKey).Result()
	if err != nil || len(phones) == 0 {
		fmt.Printf("❌ 没有找到手机号: %v\n", err)
		return
	}

	// 构建消息正文
	var summaryLines string
	for _, event := range allEvents {
		switch eventType {
		case EventCreate:
			link := fmt.Sprintf("https://hzbxtx.atlassian.net/browse/%s?linkSource=email", event["summaryKeyID"])
			summaryLines += fmt.Sprintf("- **摘要名称**: [%s](%s)\n", event["summary"], link)

		case EventDelete:
			summaryLines += fmt.Sprintf("- **摘要名称**: ~~%s %s~~\n", event["summaryKeyID"], event["summary"])

		case EventUpdateStatus:
			link := fmt.Sprintf("https://hzbxtx.atlassian.net/browse/%s?linkSource=email", event["summaryKeyID"])
			staChg := fmt.Sprintf(" **状态**: ~~%s~~ → **%s**", event["statusFrom"], event["statusTo"])
			summaryLines += fmt.Sprintf("- **摘要名称**: [%s](%s)\n- %s\n", event["summary"], link, staChg)

		case EventUpdateReport:
			link := fmt.Sprintf("https://hzbxtx.atlassian.net/browse/%s?linkSource=email", event["summaryKeyID"])
			rptChg := fmt.Sprintf("**报告人**: ~~%s~~ → **%s**", event["rptFrom"], event["rptTo"])
			summaryLines += fmt.Sprintf("- **摘要名称**: [%s](%s)\n- %s\n", event["summary"], link, rptChg)

		case EventUpdateAssigner:
			link := fmt.Sprintf("https://hzbxtx.atlassian.net/browse/%s?linkSource=email", event["summaryKeyID"])
			assChg := fmt.Sprintf("**经办人**: %s", event["assignerFromTo"])
			summaryLines += fmt.Sprintf("- **摘要名称**: [%s](%s)\n- %s\n", event["summary"], link, assChg)
		}
	}

	// 构造消息标题
	var title string
	switch eventType {
	case EventCreate:
		title = "新任务创建"
	case EventDelete:
		title = "任务被删除"
	case EventUpdateReport:
		title = "任务报告人变更"
	case EventUpdateAssigner:
		title = "任务经办人变更"
	case EventUpdateStatus:
		title = "任务状态变更"
	}

	mentionText := BuildAtMentions(phones...)

	content := fmt.Sprintf(`
### **事件通知: %s**             
%s
- **操作人**: %s
---
%s`, title, summaryLines, operator, mentionText)

	// 发送钉钉通知
	err = sendDingTalkNotification(content, toStringSlice(phones))
	if err != nil {
		fmt.Printf("⚠️ 钉钉通知发送失败: %v\n", err)
	}

	// 清理 Redis 和 Timer Map
	_ = RedisClient.Del(ctx, summaryKey).Err()
	_ = RedisClient.Del(ctx, phoneKey).Err()

	timerLock.Lock()
	delete(debounceTimers, fmt.Sprintf("timer_key:%s:%s", eventType, operator))
	timerLock.Unlock()
}

func BuildAtMentions(mobiles ...string) string {
	var mentions string
	for _, mobile := range mobiles {
		if mobile != "" {
			mentions += fmt.Sprintf("@%s ", mobile)
		}
	}
	return mentions
}

func toStringSlice(items []string) []string {
	return items
}

func WithMySQL(config *conf.MySQLConfig, fn func(db *sql.DB) error) error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=true&loc=Local",
		config.User,
		config.Password,
		config.Host,
		config.Port,
		config.Database,
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("打开数据库失败: %v", err)
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {

		}
	}(db)

	if err := db.Ping(); err != nil {
		return fmt.Errorf("数据库 Ping 失败: %v", err)
	}

	return fn(db)
}
