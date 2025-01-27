package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger interface that can be implemented by any logger
type Logger interface {
	Info(msg string, keysAndValues ...interface{})
	Error(msg string, keysAndValues ...interface{})
	Fatal(msg string, keysAndValues ...interface{})
	Debug(msg string, keysAndValues ...interface{})
	With(keysAndValues ...interface{}) Logger
}

// zapLogger implements Logger interface
type zapLogger struct {
	logger *zap.SugaredLogger
}

// NewLogger creates a new logger instance
func NewLogger(env string) Logger {
	var config zap.Config

	if env == "production" {
		config = zap.NewProductionConfig()
		config.EncoderConfig.TimeKey = "timestamp"
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	} else {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	baseLogger, err := config.Build()
	if err != nil {
		panic("failed to initialize logger: " + err.Error())
	}

	return &zapLogger{
		logger: baseLogger.Sugar(),
	}
}

func (l *zapLogger) Info(msg string, keysAndValues ...interface{}) {
	l.logger.Infow(msg, keysAndValues...)
}

func (l *zapLogger) Error(msg string, keysAndValues ...interface{}) {
	l.logger.Errorw(msg, keysAndValues...)
}

func (l *zapLogger) Fatal(msg string, keysAndValues ...interface{}) {
	l.logger.Fatalw(msg, keysAndValues...)
	os.Exit(1)
}

func (l *zapLogger) Debug(msg string, keysAndValues ...interface{}) {
	l.logger.Debugw(msg, keysAndValues...)
}

func (l *zapLogger) With(keysAndValues ...interface{}) Logger {
	return &zapLogger{
		logger: l.logger.With(keysAndValues...),
	}
}

//to be removed later
// Example of using the logger in other parts of your application:
// func someFunction(log logger.Logger) {
//     // Add request ID or other context to logs
//     reqLogger := log.With(
//         "requestID", "123",
//         "userID", "456",
//     )

//     reqLogger.Info("Processing request",
//         "method", "GET",
//         "path", "/api/trades",
//     )

//     // If something goes wrong
//     if err := doSomething(); err != nil {
//         reqLogger.Error("Failed to process request",
//             "error", err,
//             "status", 500,
//         )
//     }
// }
