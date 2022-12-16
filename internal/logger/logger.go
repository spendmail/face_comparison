package logger

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	DEBUG = "debug"
	INFO  = "info"
	WARN  = "warn"
	ERROR = "error"
)

type Config interface {
	GetLoggerLevel() string
	GetLoggerFile() string
}

type Logger struct {
	Logger *logrus.Logger
}

var ErrLogFileOpen = errors.New("unable to open a log file")

// New is a logger constructor.
func New(config Config) (*Logger, error) {
	logger := logrus.New()
	logger.Formatter = &logrus.TextFormatter{}

	file, err := os.OpenFile(config.GetLoggerFile(), os.O_APPEND|os.O_CREATE|os.O_RDWR, 0o666)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrLogFileOpen, config.GetLoggerFile())
	}

	logger.SetOutput(file)

	switch config.GetLoggerLevel() {
	case DEBUG:
		logger.SetLevel(logrus.DebugLevel)
	case INFO:
		logger.SetLevel(logrus.InfoLevel)
	case WARN:
		logger.SetLevel(logrus.WarnLevel)
	case ERROR:
		logger.SetLevel(logrus.ErrorLevel)
	}

	return &Logger{
		Logger: logger,
	}, nil
}

func (l *Logger) Trace(args ...interface{}) {
	l.Logger.Trace(args...)
}

func (l *Logger) Debug(args ...interface{}) {
	l.Logger.Debug(args...)
}

func (l *Logger) Info(args ...interface{}) {
	l.Logger.Info(args...)
}

func (l *Logger) Warn(args ...interface{}) {
	l.Logger.Warn(args...)
}

func (l *Logger) Error(args ...interface{}) {
	l.Logger.Error(args...)
}

func (l *Logger) Fatal(args ...interface{}) {
	l.Logger.Fatal(args...)
}

func (l *Logger) Panic(args ...interface{}) {
	l.Logger.Panic(args...)
}
