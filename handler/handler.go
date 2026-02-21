package handler

import (
	"context"
	"log"
	"strings"

	"github.com/tencent-connect/botgo/dto"
	"github.com/tencent-connect/botgo/dto/message"
	"github.com/tencent-connect/botgo/openapi"

	"qqbot/command"
)

// Processor 消息处理器
type Processor struct {
	API      openapi.OpenAPI
	Registry *command.Registry
}

// NewProcessor 创建消息处理器
func NewProcessor(api openapi.OpenAPI, registry *command.Registry) *Processor {
	return &Processor{
		API:      api,
		Registry: registry,
	}
}

// ProcessGroupMessage 处理群@机器人消息
func (p *Processor) ProcessGroupMessage(event *dto.WSPayload, data *dto.WSGroupATMessageData) error {
	// 提取纯文本内容（去掉@机器人的部分）
	input := strings.TrimSpace(message.ETLInput(data.Content))
	log.Printf("[群消息] GroupID=%s Author=%s Content=%s", data.GroupID, data.Author.ID, input)

	// 尝试匹配指令
	cmd, args := p.Registry.Match(input)

	ctx := &command.Context{
		Content:   input,
		Args:      args,
		Source:    command.SourceGroup,
		GroupData: data,
		API:       p.API,
		Ctx:       context.Background(),
	}

	var reply string
	if cmd != nil {
		reply = cmd.Handler(ctx)
	} else if kwHandler := p.Registry.MatchKeyword(input); kwHandler != nil {
		reply = kwHandler(ctx)
	}

	if reply == "" {
		return nil
	}

	// 发送群消息回复
	_, err := p.API.PostGroupMessage(context.Background(), data.GroupID, &dto.MessageToCreate{
		Content: reply,
		MsgType: dto.TextMsg,
		MsgID:   data.ID,
		MsgSeq:  1,
	})
	if err != nil {
		log.Printf("[群消息回复失败] GroupID=%s err=%v", data.GroupID, err)
		return err
	}
	return nil
}

// ProcessC2CMessage 处理私聊消息
func (p *Processor) ProcessC2CMessage(event *dto.WSPayload, data *dto.WSC2CMessageData) error {
	input := strings.TrimSpace(data.Content)
	log.Printf("[私聊消息] UserID=%s Content=%s", data.Author.ID, input)

	// 尝试匹配指令
	cmd, args := p.Registry.Match(input)

	ctx := &command.Context{
		Content: input,
		Args:    args,
		Source:  command.SourceC2C,
		C2CData: data,
		API:     p.API,
		Ctx:     context.Background(),
	}

	var reply string
	if cmd != nil {
		reply = cmd.Handler(ctx)
	} else if kwHandler := p.Registry.MatchKeyword(input); kwHandler != nil {
		reply = kwHandler(ctx)
	}

	if reply == "" {
		return nil
	}

	// 发送私聊消息回复
	_, err := p.API.PostC2CMessage(context.Background(), data.Author.ID, &dto.MessageToCreate{
		Content: reply,
		MsgType: dto.TextMsg,
		MsgID:   data.ID,
		MsgSeq:  1,
	})
	if err != nil {
		log.Printf("[私聊回复失败] UserID=%s err=%v", data.Author.ID, err)
		return err
	}
	return nil
}

// ProcessChannelMessage 处理频道@机器人消息
func (p *Processor) ProcessChannelMessage(event *dto.WSPayload, data *dto.WSATMessageData) error {
	input := strings.TrimSpace(message.ETLInput(data.Content))
	log.Printf("[频道消息] ChannelID=%s Author=%s Content=%s", data.ChannelID, data.Author.Username, input)

	// 尝试匹配指令
	cmd, args := p.Registry.Match(input)

	ctx := &command.Context{
		Content:     input,
		Args:        args,
		Source:      command.SourceChannel,
		ChannelData: data,
		API:         p.API,
		Ctx:         context.Background(),
	}

	var reply string
	if cmd != nil {
		reply = cmd.Handler(ctx)
	} else if kwHandler := p.Registry.MatchKeyword(input); kwHandler != nil {
		reply = kwHandler(ctx)
	}

	if reply == "" {
		return nil
	}

	// 发送频道消息回复
	_, err := p.API.PostMessage(context.Background(), data.ChannelID, &dto.MessageToCreate{
		Content: reply,
		MsgID:   data.ID,
	})
	if err != nil {
		log.Printf("[频道回复失败] ChannelID=%s err=%v", data.ChannelID, err)
		return err
	}
	return nil
}
