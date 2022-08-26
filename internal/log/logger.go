package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	stdoutPath  = "stdout"
	stderrPath  = "stderr"
	logfilePath = "tmp/log.txt"
)

var (
	Any      = zap.Any
	Duration = zap.Duration
	String   = zap.String
)

type Field = zap.Field

type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	Sync() error
}

func New(lvl Level, logToStd bool) (Logger, error) {
	outPaths := []string{logfilePath}
	errOutPaths := []string{stderrPath, logfilePath}
	if logToStd {
		outPaths = []string{stdoutPath}
		errOutPaths = []string{stderrPath}
	}
	return zap.Config{
		Level:    zap.NewAtomicLevelAt(levelsMapping[lvl]),
		Encoding: "json",
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey:     "msg",
			LevelKey:       "lvl",
			TimeKey:        "ts",
			NameKey:        "logger",
			CallerKey:      "caller",
			FunctionKey:    zapcore.OmitKey,
			StacktraceKey:  "stack",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.CapitalLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.StringDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
		OutputPaths:      outPaths,
		ErrorOutputPaths: errOutPaths,
	}.Build()
}
