package main

import (
	"flag"
	"github.com/reusedev/draw-hub/config"
	"github.com/reusedev/draw-hub/internal/components/mysql"
	"github.com/reusedev/draw-hub/internal/modules/logs"
	"github.com/reusedev/draw-hub/internal/modules/storage/ali"
	"github.com/reusedev/draw-hub/internal/service/http"
	"github.com/reusedev/draw-hub/internal/service/http/model"
	"github.com/reusedev/draw-hub/tools"
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
	flag.StringVar(&configPath, "config", "config.yml", "config file path")
}

func main() {
	flag.Parse()
	config.Init(tools.ReadFile(configPath))
	logs.InitLogger()
	mysql.CreateDataBase(config.GConfig.MySQL)
	mysql.InitMySQL(config.GConfig.MySQL)
	mysql.DB.AutoMigrate(&model.InputImage{}, &model.OutputImage{}, &model.Task{}, &model.TaskImage{}, &model.SupplierInvokeHistory{})
	ali.InitOSS(config.GConfig.AliOss)
	osSignal := make(chan os.Signal, 1)
	signal.Notify(osSignal, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	go func(ch chan os.Signal) {
		<-ch
		os.Exit(0)
	}(osSignal)
	http.Serve(httpPort)
}
