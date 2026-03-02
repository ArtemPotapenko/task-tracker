package logger

import (
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	*zap.SugaredLogger
}

func (l *Logger) Printf(format string, args ...any) {
	l.Infof(format, args...)
}

var (
	once   sync.Once
	sugar  *zap.SugaredLogger
	single *Logger
)

var Log = New()

func initStd() {
	cfg := zap.NewProductionConfig()
	cfg.Encoding = "console"
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	base, err := cfg.Build()
	if err != nil {
		base = zap.NewNop()
	}
	sugar = base.Sugar()
	single = &Logger{SugaredLogger: sugar}
}

func New() *Logger {
	once.Do(initStd)
	return single
}
