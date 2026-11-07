package logger

import (
	"fmt"
	"io"
	"os"
	"time"
)

type Level int

const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
)

type Logger struct {
	out io.Writer
}

func New(out io.Writer) *Logger {
	if out == nil {
		out = os.Stdout
	}
	return &Logger{out: out}
}

func (l *Logger) Info(format string, args ...any) {
	fmt.Fprintf(l.out, "[%s] [INFO]  %s\n", time.Now().Format("15:04:05"), fmt.Sprintf(format, args...))
}

func (l *Logger) Warn(format string, args ...any) {
	fmt.Fprintf(l.out, "[%s] [WARN]  %s\n", time.Now().Format("15:04:05"), fmt.Sprintf(format, args...))
}

func (l *Logger) Error(format string, args ...any) {
	fmt.Fprintf(l.out, "[%s] [ERROR] %s\n", time.Now().Format("15:04:05"), fmt.Sprintf(format, args...))
}
