package log

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/rs/xid"
	"log/slog"
	"os"
)

type loggerKeyType string

const loggerKey loggerKeyType = "Logger"

var rootLogger *slog.Logger

func Init(structured bool, level slog.Level) {
	var newLogger *slog.Logger
	if structured {
		newLogger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{AddSource: true, Level: level}))
	} else {
		newLogger = slog.New(NewPlainTextHandler(os.Stdout, slog.LevelDebug))
	}
	InitWithLogger(newLogger)
}

func InitWithLogger(logger *slog.Logger) {
	rootLogger = logger
	slog.SetDefault(logger)
}

func CreateRequestContextLogger(c *gin.Context) *slog.Logger {
	requestId := xid.New().String()
	correlationId := nvl(c.Request.Header.Get("X-Correlation-ID"), xid.New().String())
	logger := rootLogger.With(
		"RequestID", requestId,
		"CorrelationID", correlationId,
	)
	c.Set("RequestID", requestId)
	c.Set("CorrelationID", correlationId)
	c.Set(string(loggerKey), logger)
	c.Writer.Header().Set("X-Request-ID", requestId)
	c.Writer.Header().Set("X-Correlation-ID", correlationId)
	return logger
}

func Logger() *slog.Logger {
	return rootLogger
}

func CreateLoggerContext(ginContext *gin.Context) context.Context {
	var ctx = ginContext.Request.Context()
	if val, exist := ginContext.Get(string(loggerKey)); exist {
		ctx = context.WithValue(ctx, loggerKey, val.(*slog.Logger))
	} else {
		ctx = context.WithValue(ctx, loggerKey, rootLogger)
	}
	return context.WithValue(ctx, loggerKey, GetRequestContextLogger(ginContext))
}

func GetRequestContextLogger(ginContext *gin.Context) *slog.Logger {
	if val, exist := ginContext.Get(string(loggerKey)); exist {
		return val.(*slog.Logger)
	} else {
		return rootLogger
	}
}

func GetContextLogger(context context.Context) *slog.Logger {
	return context.Value(loggerKey).(*slog.Logger)
}

func nvl(str, defaultStr string) string {
	if str == "" {
		return defaultStr
	}
	return str
}
