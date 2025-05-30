package handler

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/reusedev/draw-hub/internal/modules/logs"
	"github.com/reusedev/draw-hub/internal/modules/model/chat/grok"
	"github.com/reusedev/draw-hub/internal/service/http/handler/request"
	"github.com/reusedev/draw-hub/internal/service/http/handler/response"
	"net/http"
)

type chatHandler struct {
	request *request.ChatCompletion
}

func newHandler(req *request.ChatCompletion) *chatHandler {
	return &chatHandler{
		request: req,
	}
}

func (c *chatHandler) chat() (*response.ChatCompletion, error) {
	resp := grok.DeepSearch(c.request.ToChatCommonRequest())
	if len(resp) == 0 {
		return nil, fmt.Errorf("no response found")
	}

	for _, v := range resp {
		m, err := v.Marsh()
		if err != nil {
			logs.Logger.Err(err).Msg("chat-DeepSearch")
			continue
		}
		logs.Logger.Info().Str("Chat response: ", string(m)).Msg("chat-DeepSearch")
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
