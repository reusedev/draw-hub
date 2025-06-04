package http

import (
	"github.com/gin-gonic/gin"
	"github.com/reusedev/draw-hub/internal/service/http/handler"
	"github.com/reusedev/draw-hub/internal/service/http/middleware"
	"net/http"
)

func Serve(port string) {
	e := gin.New()

	// 设置可信代理，解决安全警告
	// 包括本地回环地址和可能的反向代理地址
	trustedProxies := []string{
		"127.0.0.1",      // IPv4 本地回环
		"::1",            // IPv6 本地回环
		"10.0.0.0/8",     // 私有网络 A 类
		"172.16.0.0/12",  // 私有网络 B 类
		"192.168.0.0/16", // 私有网络 C 类
	}
	e.SetTrustedProxies(trustedProxies)

	initRouter(e)
	srv := &http.Server{
		Addr:    port,
		Handler: e,
	}
	if err := srv.ListenAndServe(); err != nil {
		panic(err)
	}
}

func initRouter(e *gin.Engine) {
	e.Use(gin.Recovery())
	e.Use(middleware.RequestLogger())
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
