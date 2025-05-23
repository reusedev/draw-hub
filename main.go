package main

import (
	"flag"
	"github.com/reusedev/draw-hub/config"
	"github.com/reusedev/draw-hub/internal/components/mysql"
	"github.com/reusedev/draw-hub/internal/modules/logs"
	"github.com/reusedev/draw-hub/internal/modules/storage/ali"
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
	flag.StringVar(&configPath, "config", "config.yml", "config file path")
}

func main() {
	flag.Parse()
	config.Init(readF(configPath))
	logs.InitLogger()
	mysql.CreateDataBase(config.GConfig.MySQL)
	mysql.InitMySQL(config.GConfig.MySQL)
	mysql.DB.AutoMigrate(&model.InputImage{}, &model.OutputImage{}, &model.Task{}, &model.TaskImage{}, &model.SupplierInvokeHistory{})
	ali.InitOSS(config.GConfig.AliOss)
	http.Serve(httpPort)
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
}

func readF(path string) []byte {
	data, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	return data
}
