package ports

type Logger interface {
	Error(args ...interface{})
	Info(args ...interface{})
}
