package logger

type Options struct {
	isPrint     bool
	isWroteFile bool
	flushSec    int32
	filePath    string
	fileName    string
}

type Option func(o Options) Options

func OpenPrint() Option {
	return func(o Options) Options {
		o.isPrint = true
		return o
	}
}

func OpenWriteFile() Option {
	return func(o Options) Options {
		o.isWroteFile = true
		return o
	}
}

func LogFilePath(path string) Option {
	return func(o Options) Options {
		o.filePath = path
		return o
	}
}

func LogFileName(name string) Option {
	return func(o Options) Options {
		o.fileName = name
		return o
	}
}

func FlushSec(sec int32) Option {
	return func(o Options) Options {
		o.flushSec = sec
		return o
	}
}
