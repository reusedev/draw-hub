package logs

import (
	"github.com/natefinch/lumberjack"
	"github.com/rs/zerolog"
)

var (
	Logger zerolog.Logger
)

func InitLogger() {
	logFile := &lumberjack.Logger{
		Filename:   "draw_hub.logs", // 日志文件路径
		MaxSize:    10,              // 单个日志文件最大大小（MB）
		MaxBackups: 5,               // 保留旧日志文件的最大数量
		MaxAge:     30,              // 日志文件保留的最大天数（30天后自动删除）
		Compress:   true,            // 是否压缩旧日志文件（gzip）
	}
	logger := zerolog.New(logFile).With().Timestamp().Logger()
	Logger = logger
}
