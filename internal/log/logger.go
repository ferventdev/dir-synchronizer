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
	Bool     = zap.Bool
	Duration = zap.Duration
	Error    = zap.Error
	Int64    = zap.Int64
	Uint64   = zap.Uint64
	Reflect  = zap.Reflect
	String   = zap.String
	Time     = zap.Time
)

func Cause(err error) Field {
	return zap.NamedError("cause", err)
}

func NilField(key string) Field {
	return Reflect(key, nil)
}

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
