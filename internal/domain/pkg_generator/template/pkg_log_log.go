package template

import "meta-egg/internal/domain/helper"

var TplPkgLogLog string = helper.PH_META_EGG_HEADER + `
package log

import (
	"fmt"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	log "github.com/sirupsen/logrus"
)

type Logger struct {
	fp     string
	funcn  string
	logger *log.Entry
}

var mlog = &Logger{
	logger: log.NewEntry(log.StandardLogger()),
}

// Fields type, used to pass to ` + "`" + `WithFields` + "`" + `.
type Fields map[string]interface{}

func init() {
	mlog.logger.Logger.SetLevel(log.InfoLevel)
	mlog.logger.Logger.SetReportCaller(false)
	mlog.logger.Logger.SetFormatter(&log.TextFormatter{
		DisableQuote:    true,
		TimestampFormat: "2006-01-02 15:04:05",
	})
}

// level 日志级别
// level: fatal, error, warn, info, debug
func SetLevel(level string) {
	l, err := log.ParseLevel(level)
	if err != nil {
		return // 设置失败，不改变默认的日志级别
	}
	mlog.logger.Logger.SetLevel(l)
}

// getCallerInfo 获取调用者的函数名，调用行
func (l *Logger) getCallerInfo() {
	// 调用链往上翻三层，找到调用函数的信息
	pc, file, line, ok := runtime.Caller(3)
	if !ok {
		return
	}
	funcName := runtime.FuncForPC(pc).Name()
	l.funcn = path.Base(funcName)

	parts := strings.Split(file, string(filepath.Separator))
	j := 3
	pickParts := make([]string, j)
	j--
	for i := len(parts) - 1; i >= 0 && j >= 0; i-- {
		pickParts[j] = parts[i]
		j--
	}
	pickParts = pickParts[j+1:]
	l.fp = fmt.Sprintf("%s:%d", strings.Join(pickParts, string(filepath.Separator)), line)
}

// print 日志处理函数
func (l *Logger) print(level log.Level, msg ...interface{}) {
	l.getCallerInfo()
	l.logger.WithFields(log.Fields{"filePath": l.fp, "func": l.funcn}).Log(level, msg...)
}

// printf 格式化日志处理
func (l *Logger) printf(level log.Level, format string, msg ...interface{}) {
	l.getCallerInfo()
	l.logger.WithFields(log.Fields{"filePath": l.fp, "func": l.funcn}).Logf(level, format, msg...)
}

func (l *Logger) Debug(args ...interface{}) {
	l.print(log.DebugLevel, args...)
}

func (l *Logger) Debugf(format string, args ...interface{}) {
	l.printf(log.DebugLevel, format, args...)
}

func (l *Logger) Info(args ...interface{}) {
	l.print(log.InfoLevel, args...)
}

func (l *Logger) Infof(format string, args ...interface{}) {
	l.printf(log.InfoLevel, format, args...)
}

func (l *Logger) Warn(args ...interface{}) {
	l.print(log.WarnLevel, args...)
}

func (l *Logger) Warnf(format string, args ...interface{}) {
	l.printf(log.WarnLevel, format, args...)
}

func (l *Logger) Error(args ...interface{}) {
	l.print(log.ErrorLevel, args...)
}

func (l *Logger) Errorf(format string, args ...interface{}) {
	l.printf(log.ErrorLevel, format, args...)
}

func (l *Logger) Fatal(args ...interface{}) {
	l.print(log.FatalLevel, args...)
	l.logger.Logger.Exit(1)
}

func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.printf(log.FatalLevel, format, args...)
	l.logger.Logger.Exit(1)
}

func (l *Logger) WithFields(fields Fields) *Logger {
	return &Logger{
		logger: l.logger.WithFields(log.Fields(fields)),
	}
}

func (l *Logger) WithError(err error) *Logger {
	return &Logger{
		logger: l.logger.WithError(err),
	}
}

func (l *Logger) WithField(key string, value interface{}) *Logger {
	return &Logger{
		logger: l.logger.WithField(key, value),
	}
}

func Debug(args ...interface{}) {
	mlog.print(log.DebugLevel, args...)
}

func Debugf(format string, args ...interface{}) {
	mlog.printf(log.DebugLevel, format, args...)
}

func Info(args ...interface{}) {
	mlog.print(log.InfoLevel, args...)
}

func Infof(format string, args ...interface{}) {
	mlog.printf(log.InfoLevel, format, args...)
}

func Warn(args ...interface{}) {
	mlog.print(log.WarnLevel, args...)
}

func Warnf(format string, args ...interface{}) {
	mlog.printf(log.WarnLevel, format, args...)
}

func Error(args ...interface{}) {
	mlog.print(log.ErrorLevel, args...)
}

func Errorf(format string, args ...interface{}) {
	mlog.printf(log.ErrorLevel, format, args...)
}

func Fatal(args ...interface{}) {
	mlog.print(log.FatalLevel, args...)
	mlog.logger.Logger.Exit(1)
}

func Fatalf(format string, args ...interface{}) {
	mlog.printf(log.FatalLevel, format, args...)
	mlog.logger.Logger.Exit(1)
}

func WithFields(fields Fields) *Logger {
	return &Logger{
		logger: mlog.logger.WithFields(log.Fields(fields)),
	}
}

func WithError(err error) *Logger {
	return &Logger{
		logger: mlog.logger.WithError(err),
	}
}

func WithField(key string, value interface{}) *Logger {
	return &Logger{
		logger: mlog.logger.WithField(key, value),
	}
}
`
