package scanner

import "time"

type Config struct {
	isPrintLog  bool
	isWroteFile bool
	goroutines  int32
	timeout     time.Duration
	typ         TypeScanner
	logFilePath string
	logFileName string
}

type Option func(conf Config) Config

func MaxGoroutines(max int32) Option {
	return func(conf Config) Config {
		conf.goroutines = max
		return conf
	}
}

func Timeout(sec int32) Option {
	return func(conf Config) Config {
		conf.timeout = time.Duration(sec) * time.Second
		return conf
	}
}

func TypeOfScanner(typ TypeScanner) Option {
	return func(conf Config) Config {
		conf.typ = typ
		return conf
	}
}

func OpenLogPrint() Option {
	return func(conf Config) Config {
		conf.isPrintLog = true
		return conf
	}
}

func OpenLogFileWrite() Option {
	return func(conf Config) Config {
		conf.isWroteFile = true
		return conf
	}
}

func LogFilePath(path string) Option {
	return func(conf Config) Config {
		conf.logFilePath = path
		return conf
	}
}

func LogFileName(name string) Option {
	return func(conf Config) Config {
		conf.logFileName = name
		return conf
	}
}
