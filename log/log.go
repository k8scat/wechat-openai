package log

import (
	"os"
	"path"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	initLogger sync.Once
	logger     *zap.Logger
)

const logFile = "logs/wechat-openai.log"

func GetLogger() *zap.Logger {
	initLogger.Do(func() {
		logCfg := zap.NewProductionConfig()

		if err := os.MkdirAll(path.Dir(logFile), os.ModePerm); err != nil {
			panic(err)
		}

		logCfg.OutputPaths = append(logCfg.OutputPaths, logFile)
		logCfg.EncoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder

		var err error
		logger, err = logCfg.Build()
		if err != nil {
			panic(err)
		}
	})
	return logger
}

func Sync() {
	if err := GetLogger().Sync(); err != nil {
		panic(err)
	}
}

func Info(msg string, fields ...zapcore.Field) {
	GetLogger().Info(msg, fields...)
}

func Warn(msg string, fields ...zapcore.Field) {
	GetLogger().Warn(msg, fields...)
}

func Error(msg string, fields ...zapcore.Field) {
	GetLogger().Error(msg, fields...)
}
