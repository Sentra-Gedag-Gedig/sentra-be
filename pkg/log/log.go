package log

import (
	"fmt"
	"golang.org/x/net/context"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"os"
	"path"
	"runtime"
	"strings"
	"sync"
	"time"

	formatter "github.com/antonfisher/nested-logrus-formatter"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

var (
	logger *logrus.Logger
	once   sync.Once
)

const RequestIDKey = "request_id"

type Fields = logrus.Fields

func NewLogger() *logrus.Logger {
	once.Do(func() {
		logger = logrus.New()
		logger.SetLevel(logrus.DebugLevel)

		logger.SetFormatter(&formatter.Formatter{
			NoColors:        false,
			TimestampFormat: "02 Jan 06 - 15:04",
			HideKeys:        false,
			CallerFirst:     true,
			CustomCallerFormatter: func(f *runtime.Frame) string {
				s := strings.Split(f.Function, ".")
				funcName := s[len(s)-1]
				return fmt.Sprintf(" \x1b[%dm[%s:%d][%s()]", 34, path.Base(f.File), f.Line, funcName)
			},
		})

		writers := []io.Writer{os.Stderr}

		appEnv := os.Getenv("APP_ENV")
		if appEnv != "test" {
			fileWriter := &lumberjack.Logger{
				Filename:   fmt.Sprintf("./storage/logs/app-%s.log", time.Now().Format("2006-01-02")),
				LocalTime:  true,
				Compress:   true,
				MaxSize:    100,
				MaxAge:     7,
				MaxBackups: 3,
			}
			writers = append(writers, fileWriter)
		}

		logger.SetOutput(io.MultiWriter(writers...))
		logger.SetReportCaller(true)
	})

	return logger
}

func Debug(fields Fields, msg string) {
	if fields == nil {
		fields = Fields{}
	}
	logger.WithFields(fields).Debug(msg)
}

func Info(fields Fields, msg string) {
	if fields == nil {
		fields = Fields{}
	}
	logger.WithFields(fields).Info(msg)
}

func Warn(fields Fields, msg string) {
	if fields == nil {
		fields = Fields{}
	}
	logger.WithFields(fields).Warn(msg)
}

func Error(fields Fields, msg string) {
	if fields == nil {
		fields = Fields{}
	}
	logger.WithFields(fields).Error(msg)
}

func ErrorWithTraceID(fields Fields, msg string) string {
	var traceID string
	if reqID, ok := fields["request_id"]; ok && reqID != "" {
		traceID = reqID.(string)
	} else {
		uuid, err := uuid.NewRandom()
		if err != nil {
			Error(Fields{
				"error": err.Error(),
			}, "[log.ErrorWithTraceID] failed to generate trace ID")
			traceID = "unknown"
		} else {
			traceID = uuid.String()
		}
	}

	if fields == nil {
		fields = Fields{}
	}

	fields["trace_id"] = traceID
	logger.WithFields(fields).Error(msg)

	return traceID
}

func Fatal(fields Fields, msg string) {
	if fields == nil {
		fields = Fields{}
	}
	logger.WithFields(fields).Fatal(msg)
}
func Panic(fields Fields, msg string) {
	if fields == nil {
		fields = Fields{}
	}
	logger.WithFields(fields).Panic(msg)
}

func WithRequestID(ctx context.Context) *logrus.Entry {
	requestID := "unknown"
	if ctx != nil {
		if id, ok := ctx.Value(RequestIDKey).(string); ok && id != "" {
			requestID = id
		}
	}

	return logger.WithField("request_id", requestID)
}
