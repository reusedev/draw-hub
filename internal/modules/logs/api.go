package logs

import (
	"io"
	"os"
	"strings"

	"github.com/natefinch/lumberjack"
	"github.com/reusedev/draw-hub/config"
	"github.com/rs/zerolog"
)

var (
	Logger zerolog.Logger
)

func InitLogger() {
	// 从配置中获取日志参数
	cfg := config.GConfig

	// 设置日志级别
	level := parseLogLevel(cfg.LogLevel)
	zerolog.SetGlobalLevel(level)

	// 配置日志输出
	var writers []io.Writer

	// 添加文件输出
	logFile := &lumberjack.Logger{
		Filename:   cfg.LogFile,       // 日志文件路径
		MaxSize:    cfg.LogMaxSize,    // 单个日志文件最大大小（MB）
		MaxBackups: cfg.LogMaxBackups, // 保留旧日志文件的最大数量
		MaxAge:     cfg.LogMaxAge,     // 日志文件保留的最大天数
		Compress:   true,              // 是否压缩旧日志文件（gzip）
	}
	writers = append(writers, logFile)

	// 如果是debug级别，同时输出到控制台（方便开发时查看）
	if level <= zerolog.DebugLevel {
		writers = append(writers, zerolog.ConsoleWriter{Out: os.Stdout})
	}

	// 创建多输出写入器
	multiWriter := io.MultiWriter(writers...)

	// 初始化logger
	Logger = zerolog.New(multiWriter).With().Timestamp().Logger()
}

// parseLogLevel 解析日志级别字符串为zerolog.Level
func parseLogLevel(levelStr string) zerolog.Level {
	switch strings.ToLower(levelStr) {
	case "trace":
		return zerolog.TraceLevel
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	case "fatal":
		return zerolog.FatalLevel
	case "panic":
		return zerolog.PanicLevel
	default:
		return zerolog.InfoLevel // 默认为info级别
	}
}
