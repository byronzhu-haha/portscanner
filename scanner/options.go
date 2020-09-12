package scanner

import "time"

type Config struct {
	goroutines int32
	timeout    time.Duration
	typ        TypeScanner
}

type Option func(conf Config) Config

func MaxGoroutines(max int32) Option {
	return func(conf Config) Config {
		conf.goroutines = max
		return conf
	}
}

func Timeout(sec time.Duration) Option {
	return func(conf Config) Config {
		conf.timeout = sec
		return conf
	}
}

func TypeOfScanner(typ TypeScanner) Option {
	return func(conf Config) Config {
		conf.typ = typ
		return conf
	}
}
