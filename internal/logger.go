package internal

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"time"
)

// LoggerWrapper 包装 zerolog.Logger 以实现 Logger 接口
type LoggerWrapper struct {
	logger zerolog.Logger
}

func (l *LoggerWrapper) Errorf(format string, v ...any) {
	l.logger.Error().Msgf(format, v...)
}

func (l *LoggerWrapper) Warnf(format string, v ...any) {
	l.logger.Warn().Msgf(format, v...)
}

func (l *LoggerWrapper) Debugf(format string, v ...any) {
	l.logger.Debug().Msgf(format, v...)
}

func InitLogger(debug bool) {
	// 设置全局时间格式
	zerolog.TimeFieldFormat = time.RFC3339
	// 设置控制台输出格式
	consoleWriter := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: "2006-01-02 15:04:05",
	}

	// 创建多写入器并启用 Caller
	log.Logger = zerolog.New(consoleWriter).
		With().
		Timestamp().
		Caller(). // 启用 Caller 信息
		Logger()
	reqLogger := &LoggerWrapper{
		logger: log.Logger,
	}

	// 根据debug参数设置日志级别
	if debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		BApi.SetDev(reqLogger)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
}
