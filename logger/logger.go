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

const (
	defaultFlushSec = 30
	writeByteSize   = 2048
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
	buf  *bytes.Buffer
	mu   sync.Mutex
}

func NewLogger(opts ...Option) Logger {
	l := &logger{}
	for _, opt := range opts {
		l.opt = opt(l.opt)
	}
	if l.opt.isWroteFile {
		if l.opt.filePath != "." {
			l.mkdir()
		} else if !strings.HasSuffix(l.opt.filePath, string(os.PathSeparator)) {
			l.opt.filePath += string(os.PathSeparator)
		}
		f, err := os.OpenFile(l.opt.filePath + l.opt.fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			panic(err)
		}
		l.buf = &bytes.Buffer{}
		l.file = f
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
	if l.opt.isWroteFile {
		l.write()
		_ = l.wr.Flush()
		_ = l.file.Close()
	}
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
		l.mu.Unlock()
		return
	}
	if l.buf.Len() >= writeByteSize {
		l.write()
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
		l.write()
		l.mu.Unlock()
	}
}

func (l *logger) write() {
	if l.buf.Len() <= 0 {
		return
	}
	_, err := l.wr.WriteString(l.buf.String())
	if err != nil {
		println(err)
	}
}

func (l *logger) mkdir() {
	f, err := os.Stat(l.opt.filePath)
	if err != nil || f.IsDir() == false {
		if err := os.Mkdir(l.opt.filePath, os.ModePerm); err != nil {
			panic("日志目录创建失败, "+err.Error())
		}
	}
}
