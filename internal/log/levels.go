package log

import (
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Level string

const (
	DebugLevel = "debug"
	InfoLevel  = "info"
	WarnLevel  = "warn"
	ErrorLevel = "error"
)

func (l Level) IsValid() bool {
	_, ok := levelsMapping[Level(strings.ToLower(string(l)))]
	return ok
}

var levelsMapping = map[Level]zapcore.Level{
	DebugLevel: zap.DebugLevel,
	InfoLevel:  zap.InfoLevel,
	WarnLevel:  zap.WarnLevel,
	ErrorLevel: zap.ErrorLevel,
}
