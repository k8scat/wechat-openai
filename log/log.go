package log

import (
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	initLogger sync.Once
	logger     *zap.Logger
)

const logFile = "wechat-openai.log"

func GetLogger() *zap.Logger {
	initLogger.Do(func() {
		logCfg := zap.NewProductionConfig()
		logCfg.OutputPaths = append(logCfg.OutputPaths, logFile)
		logCfg.EncoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			encodeTimeLayout(t, "2006-01-02 15:04:05.000", enc)
		}

		var err error
		logger, err = logCfg.Build()
		if err != nil {
			panic(err)
		}
	})
	return logger
}

func encodeTimeLayout(t time.Time, layout string, enc zapcore.PrimitiveArrayEncoder) {
	type appendTimeEncoder interface {
		AppendTimeLayout(time.Time, string)
	}

	if enc, ok := enc.(appendTimeEncoder); ok {
		enc.AppendTimeLayout(t, layout)
		return
	}

	enc.AppendString(t.Format(layout))
}

func Sync() {
	GetLogger().Sync()
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
