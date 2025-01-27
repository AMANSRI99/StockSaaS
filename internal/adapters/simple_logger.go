// filepath: /d:/code/StockSaaS/internal/adapters/simple_logger.go
package adapters

import (
    "log"
)

type SimpleLogger struct{}

func NewSimpleLogger() *SimpleLogger {
    return &SimpleLogger{}
}

func (l *SimpleLogger) Error(args ...interface{}) {
    log.Println("[ERROR]", args)
}

func (l *SimpleLogger) Info(args ...interface{}) {
    log.Println("[INFO]", args)
}