package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var instance = initialize()

type Logger interface {
	Debugf(tmpl string, args ...any)
	Infof(tmpl string, args ...any)
	Warnf(tmpl string, args ...any)
	Errorf(tmpl string, args ...any)
}

func Default() Logger {
	return instance
}

func initialize() zapLog {
	enc := zap.NewDevelopmentEncoderConfig()
	enc.EncodeLevel = zapcore.CapitalColorLevelEncoder
	encoder := zapcore.NewConsoleEncoder(enc)
	syncer := zapcore.AddSync(os.Stdout)

	core := zapcore.NewCore(encoder, syncer, zapcore.DebugLevel)
	lg := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1), zap.AddStacktrace(zapcore.ErrorLevel))

	return zapLog{sugar: lg.Sugar()}
}

type zapLog struct {
	sugar *zap.SugaredLogger
}

func (z zapLog) Debugf(tmpl string, args ...any) {
	z.sugar.Debugf(tmpl, args...)
}

func (z zapLog) Infof(tmpl string, args ...any) {
	z.sugar.Infof(tmpl, args...)
}

func (z zapLog) Warnf(tmpl string, args ...any) {
	z.sugar.Warnf(tmpl, args...)
}

func (z zapLog) Errorf(tmpl string, args ...any) {
	z.sugar.Errorf(tmpl, args...)
}
