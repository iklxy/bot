package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/tencent-connect/botgo"
	"github.com/tencent-connect/botgo/dto"
	"github.com/tencent-connect/botgo/event"
	"github.com/tencent-connect/botgo/interaction/webhook"
	"github.com/tencent-connect/botgo/token"

	"qqbot/command"
	"qqbot/config"
	"qqbot/handler"
)

func main() {
	// 1. 加载配置
	cfg := config.Load("config.yaml")
	log.Printf("[启动] AppID=%s Debug=%v", cfg.AppID, cfg.Debug)

	// 2. 创建凭证与 token
	credentials := &token.QQBotCredentials{
		AppID:     cfg.AppID,
		AppSecret: cfg.AppSecret,
	}
	tokenSource := token.NewQQBotTokenSource(credentials)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := token.StartRefreshAccessToken(ctx, tokenSource); err != nil {
		log.Fatalln("刷新 AccessToken 失败:", err)
	}

	// 3. 初始化 OpenAPI
	api := botgo.NewOpenAPI(credentials.AppID, tokenSource).
		WithTimeout(5 * time.Second).
		SetDebug(cfg.Debug)

	// 4. 初始化指令注册中心并注册内置指令
	registry := command.NewRegistry("/")
	registerBuiltinCommands(registry)

	// 5. 创建消息处理器
	proc := handler.NewProcessor(api, registry)

	// 6. 注册事件处理函数
	_ = event.RegisterHandlers(
		// 群@机器人消息
		groupATMessageHandler(proc),
		// C2C 私聊消息
		c2cMessageHandler(proc),
		// 频道@机器人消息
		channelATMessageHandler(proc),
	)

	// 7. 启动 HTTP 服务（Webhook 回调）
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	http.HandleFunc(cfg.Path, func(w http.ResponseWriter, r *http.Request) {
		webhook.HTTPHandler(w, r, credentials)
	})

	log.Printf("[启动] HTTP 服务监听 %s%s", addr, cfg.Path)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal("HTTP 服务启动失败:", err)
	}
}

// ======================== 事件处理函数 ========================

func groupATMessageHandler(proc *handler.Processor) event.GroupATMessageEventHandler {
	return func(payload *dto.WSPayload, data *dto.WSGroupATMessageData) error {
		return proc.ProcessGroupMessage(payload, data)
	}
}

func c2cMessageHandler(proc *handler.Processor) event.C2CMessageEventHandler {
	return func(payload *dto.WSPayload, data *dto.WSC2CMessageData) error {
		return proc.ProcessC2CMessage(payload, data)
	}
}

func channelATMessageHandler(proc *handler.Processor) event.ATMessageEventHandler {
	return func(payload *dto.WSPayload, data *dto.WSATMessageData) error {
		return proc.ProcessChannelMessage(payload, data)
	}
}

// ======================== 内置指令注册 ========================

func registerBuiltinCommands(r *command.Registry) {
	// /help - 显示帮助（通过闭包引用 registry）
	r.Register(&command.Cmd{
		Name:        "help",
		Description: "显示可用指令列表",
		Handler: func(ctx *command.Context) string {
			return r.HelpText()
		},
	})

	// /ping - 测试机器人是否在线
	r.Register(&command.Cmd{
		Name:        "ping",
		Description: "测试机器人是否在线",
		Handler: func(ctx *command.Context) string {
			return "pong!"
		},
	})

	// /morning - 早安播报
	r.Register(&command.Cmd{
		Name:        "morning",
		Description: "早安问候，播报当前时间",
		Handler: func(ctx *command.Context) string {
			now := time.Now().Format("2006年01月02日 15:04:05")
			return fmt.Sprintf("早上好！现在是 %s，祝你今天元气满满！", now)
		},
	})

	// /night - 晚安播报
	r.Register(&command.Cmd{
		Name:        "night",
		Description: "晚安问候，播报当前时间",
		Handler: func(ctx *command.Context) string {
			now := time.Now().Format("2006年01月02日 15:04:05")
			return fmt.Sprintf("晚安！现在是 %s，愿你好梦，明天见！", now)
		},
	})

	// /lucky - 今日幸运指数
	r.Register(&command.Cmd{
		Name:        "lucky",
		Description: "查看今日幸运指数",
		Handler: func(ctx *command.Context) string {
			score := rand.Intn(101) // 0~100
			var comment string
			switch {
			case score >= 90:
				comment = "运气爆棚！今天买彩票说不定能中！"
			case score >= 70:
				comment = "运气不错，适合尝试新事物！"
			case score >= 50:
				comment = "运气一般般，平平淡淡也是福。"
			case score >= 30:
				comment = "运气稍差，小心行事为妙。"
			default:
				comment = "今天运气不太行...不如躺平休息一下？"
			}
			return fmt.Sprintf("你的今日幸运指数: %d/100\n%s", score, comment)
		},
	})

	// ======================== 关键词注册 ========================

	// 询问"谁创造了你"
	r.RegisterKeyword(&command.Keyword{
		Contains: []string{"谁创造了你", "谁做的你", "你的创造者", "谁开发的你", "你是谁做的", "你的作者"},
		Handler: func(ctx *command.Context) string {
			return "我是由 lxy 创造的！"
		},
	})
}
