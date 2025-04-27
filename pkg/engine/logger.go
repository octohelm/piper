package engine

import (
	"context"

	"github.com/go-courier/logr"
	"github.com/innoai-tech/infra/pkg/configuration"
	"github.com/octohelm/piper/internal/logger"
)

// +gengo:enum
type LogLevel string

const (
	ErrorLevel LogLevel = "error"
	WarnLevel  LogLevel = "warn"
	InfoLevel  LogLevel = "info"
	DebugLevel LogLevel = "debug"
)

type Logger struct {
	// Log level
	LogLevel LogLevel `flag:",omitzero"`
	logger   logr.Logger
}

func (l *Logger) SetDefaults() {
	if l.LogLevel == "" {
		l.LogLevel = InfoLevel
	}
}

func (l *Logger) Init(ctx context.Context) error {
	if l.logger == nil {
		lvl, _ := logr.ParseLevel(string(l.LogLevel))
		l.logger = &logger.Logger{Enabled: lvl}
	}
	return nil
}

func (l *Logger) InjectContext(ctx context.Context) context.Context {
	return configuration.InjectContext(
		ctx,
		configuration.InjectContextFunc(logr.WithLogger, l.logger),
	)
}
