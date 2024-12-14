package util

import (
	"fmt"
	"log"
	"runtime/debug"
)

type LogLevel int

const (
	Debug LogLevel = iota
	Trace
	Information
	Warning
	Error
	Critical
	None
)

type Logger struct {
	Level LogLevel
}

func (l *Logger) logPrint(level LogLevel, name string, v ...any) {
	if level >= l.Level {
		log.Println(fmt.Sprintf("[%s]", name), v)
	}
}

func (l *Logger) LogDebug(v ...any) {
	l.logPrint(Debug, "Debug", v)
}

func (l *Logger) LogTrace(v ...any) {
	l.logPrint(Trace, "Trace", v)
}

func (l *Logger) LogInformation(v ...any) {
	l.logPrint(Information, "Information", v)
}

func (l *Logger) LogWarning(v ...any) {
	l.logPrint(Warning, "Warning", v)
}

func (l *Logger) LogError(v ...any) {
	l.logPrint(Error, "Error", v)
}

func (l *Logger) LogCritical(v ...any) {
	l.logPrint(Critical, "Critical", v)
}

var Log Logger

// 初始化函数
func init() {
	Log.Level = Debug
}

func LogDebug(v ...any) {
	Log.LogDebug(v)
}

func LogTrace(v ...any) {
	Log.LogTrace(v)
}

func LogInformation(v ...any) {
	Log.LogInformation(v)
}

func LogWarning(v ...any) {
	Log.LogWarning(v)
}

func LogError(v ...any) {
	Log.LogError(v)
	if isDebug {
		log.Println("--------------------------------------------------")
		debug.PrintStack()
		log.Println("--------------------------------------------------")
	}
}

func LogCritical(v ...any) {
	Log.LogCritical(v)
	if isDebug {
		log.Println("==================================================")
		debug.PrintStack()
		log.Println("==================================================")
	}
}
