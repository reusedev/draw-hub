package request

import "github.com/reusedev/draw-hub/internal/modules/ai/chat"

type ChatRequest interface {
	ToChatCommonRequest() chat.CommonRequest
}

type ChatCompletion struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type ChatCompletionV2 struct {
	Model    string      `json:"model"`
	Messages []MessageV2 `json:"messages"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type MessageV2 struct {
	Role    string    `json:"role"`
	Content []Content `json:"content"`
}

type Content struct {
	Type     string    `json:"type"`
	Text     string    `json:"text,omitempty"`
	ImageURL *ImageURL `json:"image_url,omitempty"`
}

type ImageURL struct {
	URL string `json:"url"`
}

func (c *ChatCompletion) ToChatCommonRequest() chat.CommonRequest {
	ret := chat.CommonRequest{
		Model:    c.Model,
		Messages: make([]chat.Message, len(c.Messages)),
	}
	for i, msg := range c.Messages {
		ret.Messages[i] = chat.Message{
			Role: msg.Role,
			Content: []chat.Content{
				{
					Type: "text",
					Text: msg.Content,
				},
			},
		}
	}
	return ret
}

func (c *ChatCompletionV2) ToChatCommonRequest() chat.CommonRequest {
	ret := chat.CommonRequest{
		Model:    c.Model,
		Messages: make([]chat.Message, len(c.Messages)),
	}
	for i, msg := range c.Messages {
		ret.Messages[i] = chat.Message{
			Role:    msg.Role,
			Content: make([]chat.Content, len(msg.Content)),
		}
		for j, content := range msg.Content {
			ret.Messages[i].Content[j] = chat.Content{
				Type: content.Type,
				Text: content.Text,
			}
			if content.ImageURL != nil {
				ret.Messages[i].Content[j].ImageURL = chat.ImageURL{URL: content.ImageURL.URL}
			}
		}
	}
	return ret
}
