package main

import (
	"flag"
	"github.com/reusedev/draw-hub/config"
	"github.com/reusedev/draw-hub/internal/components/mysql"
	"github.com/reusedev/draw-hub/internal/service/http"
	"github.com/reusedev/draw-hub/internal/service/http/model"
	"os"
	"os/signal"
	"syscall"
)

var (
	httpPort   string
	configPath string
)

func init() {
	flag.StringVar(&httpPort, "http-port", ":80", "listen http port")
	flag.StringVar(&configPath, "config", "config.yaml", "config file path")
}

func main() {
	flag.Parse()
	config.Init(configPath)
	mysql.InitMySQL(config.GConfig.MySQL)
	mysql.DB.AutoMigrate(&model.InputImage{}, &model.OutputImage{})
	http.Serve(httpPort)
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
}
