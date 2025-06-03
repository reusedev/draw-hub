package request

import "github.com/reusedev/draw-hub/internal/modules/ai/chat"

type ChatCompletion struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func (c *ChatCompletion) ToChatCommonRequest() chat.CommonRequest {
	ret := chat.CommonRequest{
		Model:    c.Model,
		Messages: make([]chat.Message, len(c.Messages)),
	}
	for i, msg := range c.Messages {
		ret.Messages[i] = chat.Message{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}
	return ret
}
