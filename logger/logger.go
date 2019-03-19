package logger

type LeveledLogger interface {
	Debugf(format string, v ...interface{})
	Debug(v ...interface{})
	Debugln(v ...interface{})
	Infof(format string, v ...interface{})
	Info(v ...interface{})
	Infoln(v ...interface{})
	Warnf(format string, v ...interface{})
	Warn(v ...interface{})
	Warnln(v ...interface{})
	Errorf(format string, v ...interface{})
	Error(v ...interface{})
	Errorln(v ...interface{})
}
