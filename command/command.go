package command

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/tencent-connect/botgo/dto"
	"github.com/tencent-connect/botgo/openapi"
)

// Handler 指令处理函数签名
// 返回要回复的文本内容，如果返回空字符串则不回复
type Handler func(ctx *Context) string

// Context 指令执行上下文
type Context struct {
	// 原始消息内容（已去除指令前缀）
	Content string
	// 指令参数（按空格分割）
	Args []string
	// 消息来源类型
	Source SourceType
	// 群消息数据（仅群消息时有值）
	GroupData *dto.WSGroupATMessageData
	// C2C 消息数据（仅私聊时有值）
	C2CData *dto.WSC2CMessageData
	// 频道消息数据（仅频道时有值）
	ChannelData *dto.WSATMessageData
	// OpenAPI 实例
	API openapi.OpenAPI
	// Go context
	Ctx context.Context
}

// SourceType 消息来源类型
type SourceType int

const (
	SourceGroup   SourceType = iota // 群消息
	SourceC2C                       // 私聊消息
	SourceChannel                   // 频道消息
)

// Cmd 指令定义
type Cmd struct {
	Name        string  // 指令名（不含前缀）
	Description string  // 指令描述
	Handler     Handler // 处理函数
}

// Registry 指令注册中心
type Registry struct {
	mu       sync.RWMutex
	commands map[string]*Cmd
	keywords []*Keyword
	prefix   string // 指令前缀，默认 "/"
}

// Keyword 关键词触发规则
type Keyword struct {
	Contains []string // 消息中包含这些关键词之一即触发
	Handler  Handler  // 处理函数
}

// NewRegistry 创建指令注册中心
func NewRegistry(prefix string) *Registry {
	if prefix == "" {
		prefix = "/"
	}
	return &Registry{
		commands: make(map[string]*Cmd),
		prefix:   prefix,
	}
}

// Register 注册一个指令
func (r *Registry) Register(cmd *Cmd) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.commands[cmd.Name] = cmd
	log.Printf("[指令注册] %s%s - %s", r.prefix, cmd.Name, cmd.Description)
}

// Match 匹配输入文本，返回对应的指令和参数
// 如果没有匹配到指令，返回 nil
func (r *Registry) Match(input string) (*Cmd, []string) {
	input = strings.TrimSpace(input)

	// 检查是否以指令前缀开头
	if !strings.HasPrefix(input, r.prefix) {
		return nil, nil
	}

	// 去掉前缀
	input = strings.TrimPrefix(input, r.prefix)
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return nil, nil
	}

	cmdName := strings.ToLower(parts[0])
	args := parts[1:]

	r.mu.RLock()
	defer r.mu.RUnlock()

	if cmd, ok := r.commands[cmdName]; ok {
		return cmd, args
	}
	return nil, nil
}

// List 列出所有已注册指令
func (r *Registry) List() []*Cmd {
	r.mu.RLock()
	defer r.mu.RUnlock()

	cmds := make([]*Cmd, 0, len(r.commands))
	for _, cmd := range r.commands {
		cmds = append(cmds, cmd)
	}
	return cmds
}

// RegisterKeyword 注册关键词触发规则
func (r *Registry) RegisterKeyword(kw *Keyword) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.keywords = append(r.keywords, kw)
	log.Printf("[关键词注册] 触发词=%v", kw.Contains)
}

// MatchKeyword 匹配关键词，返回对应的处理函数
func (r *Registry) MatchKeyword(input string) Handler {
	r.mu.RLock()
	defer r.mu.RUnlock()

	lower := strings.ToLower(input)
	for _, kw := range r.keywords {
		for _, word := range kw.Contains {
			if strings.Contains(lower, strings.ToLower(word)) {
				return kw.Handler
			}
		}
	}
	return nil
}

// HelpText 生成帮助文本
func (r *Registry) HelpText() string {
	cmds := r.List()
	if len(cmds) == 0 {
		return "暂无可用指令"
	}

	var sb strings.Builder
	sb.WriteString("可用指令列表:\n")
	for _, cmd := range cmds {
		sb.WriteString(fmt.Sprintf("  %s%s - %s\n", r.prefix, cmd.Name, cmd.Description))
	}
	return sb.String()
}
