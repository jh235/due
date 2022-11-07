package log

var globalLogger Logger

func init() {
	SetLogger(NewLogger(WithCallerSkip(1)))
}

// SetLogger 设置日志记录器
func SetLogger(logger Logger) {
	globalLogger = logger
}

// GetLogger 获取日志记录器
func GetLogger() Logger {
	return globalLogger
}

// Debug 打印调试日志
func Debug(a ...interface{}) {
	globalLogger.Debug(a...)
}

// Debugf 打印调试模板日志
func Debugf(format string, a ...interface{}) {
	globalLogger.Debugf(format, a...)
}

// Info 打印信息日志
func Info(a ...interface{}) {
	globalLogger.Info(a...)
}

// Infof 打印信息模板日志
func Infof(format string, a ...interface{}) {
	globalLogger.Infof(format, a...)
}

// Warn 打印警告日志
func Warn(a ...interface{}) {
	globalLogger.Warn(a...)
}

// Warnf 打印警告模板日志
func Warnf(format string, a ...interface{}) {
	globalLogger.Warnf(format, a...)
}

// Error 打印错误日志
func Error(a ...interface{}) {
	globalLogger.Error(a...)
}

// Errorf 打印错误模板日志
func Errorf(format string, a ...interface{}) {
	globalLogger.Errorf(format, a...)
}

// Fatal 打印致命错误日志
func Fatal(a ...interface{}) {
	globalLogger.Fatal(a...)
}

// Fatalf 打印致命错误模板日志
func Fatalf(format string, a ...interface{}) {
	globalLogger.Fatalf(format, a...)
}

// Panic 打印Panic日志
func Panic(a ...interface{}) {
	globalLogger.Panic(a...)
}

// Panicf 打印Panic模板日志
func Panicf(format string, a ...interface{}) {
	GetLogger().Panicf(format, a...)
}
