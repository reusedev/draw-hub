package main

import (
	"context"
	"flag"
	"github.com/reusedev/draw-hub/config"
	"github.com/reusedev/draw-hub/internal/components/mysql"
	"github.com/reusedev/draw-hub/internal/modules/logs"
	"github.com/reusedev/draw-hub/internal/modules/model"
	"github.com/reusedev/draw-hub/internal/modules/queue"
	"github.com/reusedev/draw-hub/internal/modules/storage/ali"
	"github.com/reusedev/draw-hub/internal/service/http"
	"github.com/reusedev/draw-hub/internal/service/http/handler"
	"github.com/reusedev/draw-hub/tools"
	"os"
	"os/signal"
	"sync"
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
	ctx, cancel := context.WithCancel(context.Background())
	flag.Parse()
	config.Init(tools.PanicOnError(tools.ReadFile(configPath)))
	config.InitTokenManager(ctx)
	logs.InitLogger()
	syscall.Umask(0007)
	wg := &sync.WaitGroup{}
	queue.InitImageTaskQueue(ctx, wg)
	mysql.CreateDataBase(config.GConfig.MySQL)
	mysql.InitMySQL(config.GConfig.MySQL)
	mysql.DB.AutoMigrate(&model.InputImage{}, &model.OutputImage{}, &model.Task{}, &model.TaskImage{}, &model.SupplierInvokeHistory{})
	mysql.FieldMigrate()
	ali.InitOSS(config.GConfig.AliOss)
	ali.InitOSSSg(config.GConfig.AliOssSg)
	handler.EnqueueUnfinishedTask()
	osSignal := make(chan os.Signal, 1)
	signal.Notify(osSignal, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	go func(ch chan os.Signal) {
		<-ch
		cancel()
		wg.Wait()
		os.Exit(0)
	}(osSignal)
	http.Serve(httpPort)
}
