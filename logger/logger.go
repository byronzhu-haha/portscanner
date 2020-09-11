package logger

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

type Logger interface {
	Printf(format string, v ...interface{})
	Infof(format string, v ...interface{})
	Debugf(format string, v ...interface{})
	Warnf(format string, v ...interface{})
	Errorf(format string, v ...interface{})
	Close()
}

type logger struct {
	opt  Options
	file *os.File
	wr   *bufio.Writer
	buf  bytes.Buffer
	mu   sync.Mutex
}

func NewLogger(opts ...Option) Logger {
	l := &logger{}
	for _, opt := range opts {
		l.opt = opt(l.opt)
	}
	if l.opt.isWroteFile {
		f, err := os.Create(l.opt.filePath + "/" + l.opt.fileName)
		if err != nil {
			panic(err)
		}
		l.file = f
		if l.opt.flushSec <= 0 {
			l.opt.flushSec = 30
		}
		l.wr = bufio.NewWriter(f)
		go l.backgroundWrite()
	}
	return l
}

func (l *logger) Printf(format string, v ...interface{}) {
	l.print("", format, v...)
}

func (l *logger) Infof(format string, v ...interface{}) {
	l.print("[INFO]", format, v...)
}

func (l *logger) Debugf(format string, v ...interface{}) {
	l.print("[DEBUG]", format, v...)
}

func (l *logger) Warnf(format string, v ...interface{}) {
	l.print("[WARN]", format, v...)
}

func (l *logger) Errorf(format string, v ...interface{}) {
	l.print("[ERROR]", format, v...)
}

func (l *logger) Close() {
	_ = l.file.Close()
	_ = l.wr.Flush()
}

func (l *logger) print(head, format string, v ...interface{}) {
	if !l.opt.isPrint {
		return
	}

	format = fmt.Sprintf(format, v...)
	if head != "" {
		format = fmt.Sprintf("%s %s: %s", head, l.now(), format)
	}
	println(format)

	if !l.opt.isWroteFile {
		return
	}
	if !strings.HasSuffix(format, "\n") {
		format += "\n"
	}
	l.mu.Lock()
	_, err := l.buf.WriteString(format)
	if err != nil {
		println(err)
	}
	l.mu.Unlock()
}

func (l *logger) now() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func (l *logger) backgroundWrite() {
	t := time.NewTicker(time.Duration(l.opt.flushSec) * time.Second)
	for range t.C {
		l.mu.Lock()
		if l.buf.Len() != 0 {
			_, _  = l.buf.WriteTo(l.wr)
		}
		l.mu.Unlock()
	}
}
