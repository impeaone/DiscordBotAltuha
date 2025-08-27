package logger

type LoggerPrefixes interface {
	Info(string, string)
	Warning(string, string)
	Error(string, string)
}
