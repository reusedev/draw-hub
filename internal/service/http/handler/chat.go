package handler

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/reusedev/draw-hub/internal/modules/ai/chat"
	"github.com/reusedev/draw-hub/internal/modules/ai/chat/common"
	"github.com/reusedev/draw-hub/internal/modules/logs"
	"github.com/reusedev/draw-hub/internal/service/http/handler/request"
	"github.com/reusedev/draw-hub/internal/service/http/handler/response"
	"net/http"
)

type chatHandler struct {
	request request.ChatRequest
}

func newHandler(req request.ChatRequest) *chatHandler {
	return &chatHandler{
		request: req,
	}
}

func (c *chatHandler) chat() (*response.ChatCompletion, error) {
	resp := common.Chat(c.request.ToChatCommonRequest())
	if len(resp) == 0 {
		return nil, fmt.Errorf("no response found")
	}

	for _, v := range resp {
		m, err := v.Marsh()
		if err != nil {
			logs.Logger.Err(err).Msg("chat")
			continue
		}
		logs.Logger.Info().Str("Chat response: ", string(m)).Msg("chat")
		if v.Succeed() {
			ret := &response.ChatCompletion{}
			err = json.Unmarshal([]byte(v.RawBody()), ret)
			if err != nil {
				return nil, err
			}
			return ret, nil
		}
	}
	return nil, fmt.Errorf("no successful response found")
}

func ChatCompletions(c *gin.Context) {
	req := &request.ChatCompletion{}
	err := c.ShouldBindJSON(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ParamError)
		return
	}
	handler := newHandler(req)
	resp, err := handler.chat()
	if err != nil {
		logs.Logger.Err(err).Msg("chat-ChatCompletions")
		c.JSON(http.StatusInternalServerError, response.InternalError)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func ChatCompletionsV2(c *gin.Context) {
	req := &request.ChatCompletionV2{}
	err := c.ShouldBindJSON(req)
	if err != nil {
		fmt.Println(err.Error())
		c.JSON(http.StatusBadRequest, response.ParamError)
		return
	}
	handler := newHandler(req)
	resp, err := handler.chat()
	if err != nil {
		logs.Logger.Err(err).Msg("chat-ChatCompletions")
		c.JSON(http.StatusInternalServerError, response.InternalError)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// Image2Text 图生文 POST /api/v2/image2text
// 接收图片 URL 和提示词，返回 AI 生成的文本描述
func Image2Text(c *gin.Context) {
	var req struct {
		URL    string `json:"url" binding:"required"`
		Prompt string `json:"prompt" binding:"required"`
		Model  string `json:"model"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ParamError)
		return
	}

	// 默认模型
	if req.Model == "" {
		req.Model = "gpt-5"
	}

	// 构造 chat CommonRequest，包含 image_url + text 两种 content
	chatReq := chat.CommonRequest{
		Model: req.Model,
		Messages: []chat.Message{
			{
				Role: "user",
				Content: []chat.Content{
					{
						Type:     "image_url",
						ImageURL: chat.ImageURL{URL: req.URL},
					},
					{
						Type: "text",
						Text: req.Prompt,
					},
				},
			},
		},
	}

	resp := common.Chat(chatReq)
	if len(resp) == 0 {
		logs.Logger.Error().Msg("Image2Text: no response")
		c.JSON(http.StatusInternalServerError, response.InternalError)
		return
	}

	for _, v := range resp {
		if v.Succeed() {
			chatCompletion := &response.ChatCompletion{}
			err := json.Unmarshal([]byte(v.RawBody()), chatCompletion)
			if err != nil {
				logs.Logger.Err(err).Msg("Image2Text-Unmarshal")
				c.JSON(http.StatusInternalServerError, response.InternalError)
				return
			}
			content := ""
			if len(chatCompletion.Choices) > 0 {
				content = chatCompletion.Choices[0].Message.Content
			}
			c.JSON(http.StatusOK, response.SuccessWithData(gin.H{
				"content": content,
			}))
			return
		}
	}

	logs.Logger.Error().Msg("Image2Text: no successful response")
	c.JSON(http.StatusInternalServerError, response.InternalError)
}
