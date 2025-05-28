package http

import (
	"github.com/gin-gonic/gin"
	"github.com/reusedev/draw-hub/internal/service/http/handler"
)

func Serve(port string) {
	e := gin.New()
	initRouter(e)
	if err := e.Run(port); err != nil {
		panic(err)
	}
}

func initRouter(e *gin.Engine) {
	e.Use(gin.Recovery())
	v1 := e.Group("/v1")
	task := v1.Group("/task")
	{
		task.POST("/slow", handler.SlowSpeed)
		task.POST("/fast", handler.FastSpeed)
		task.GET("", handler.TaskQuery)
	}
	chat := v1.Group("/chat")
	{
		chat.POST("/completions", handler.ChatCompletions)
	}

	file := v1.Group("/images")
	{
		file.POST("", handler.UploadImage)
		file.GET("", handler.GetImage)
	}
}
