package log

import (
	"os"

	"github.com/rs/zerolog"
)

type Level string

const (
	LevelDebug   Level = "debug"
	LevelInfo    Level = "info"
	LevelWarning Level = "warning"
	LevelError   Level = "error"
)

type Logger struct {
	zerolog.Logger
}

func New(level Level) Logger {
	zerolog.TimeFieldFormat = "02-01-2006 15:04:05"
	zerolog.LevelWarnValue = "warning"

	isNotParsed := false
	var logLevel zerolog.Level
	switch level {
	case LevelDebug:
		logLevel = zerolog.DebugLevel
	case LevelInfo:
		logLevel = zerolog.InfoLevel
	case LevelWarning:
		logLevel = zerolog.WarnLevel
	case LevelError:
		logLevel = zerolog.ErrorLevel
	default:
		isNotParsed = true
		logLevel = zerolog.InfoLevel
	}

	logger := zerolog.New(os.Stdout).
		Level(logLevel).
		With().
		Timestamp().
		Logger()

	if isNotParsed {
		logger.Warn().Msgf("loglevel %s not parsed, defaulting to info", level)
	}

	return Logger{
		Logger: logger,
	}
}
