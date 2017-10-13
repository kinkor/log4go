package log4go

import (
	"sync"
	"time"
	"strings"
	"runtime"
	"fmt"
	"errors"
	"bytes"
)

/* Log4go version number constant */
const (
	LOG4G_VERSION = "v0.1.0"
)

/* Log level constant */
const (
	DEBUG = iota
	INFO
	WARN
	ERROR
)

/* Log level string array */
var (
	levelStrings    = [...]string{"DEBUG", "INFO", "WARN", "ERROR"}
	strToLevel      = map[string]Level{"DEBUG": DEBUG, "INFO": INFO, "WARN": WARN, "ERROR": ERROR}
	logBufferLength = 32
)

/* Log level type */
type Level int

/* Level to string */
func (l Level) String() string {
	if l < 0 || int(l) > len(levelStrings) {
		return ""
	}
	
	return levelStrings[int(l)]
}

/* Log record type */
type Record struct {
	Level
	Content string
	Size    int
	Time    time.Time
	Source  string
}

/* Log writer interface */
type Writer interface {
	Write(r *Record)
	Close()
}

/* Logger type */
type Logger struct {
	*Config
	Writers map[string]Writer
}

/* Logger single instance */
var logger *Logger
var once sync.Once
var wg sync.WaitGroup // 用于控制缓存的同步 防止数据丢失

//创建一个logger单例
func NewLogger(configPath string) *Logger {
	once.Do(func() {
		logger = &Logger{}
		logger.Config = new(Config)
		logger.Config.InitConfig(configPath)
		logger.Writers = make(map[string]Writer)
		s := strings.Split(logger.Item("log4go.type"), ",")
		for i := 0; i < len(s); i++ {
			switch strings.TrimSpace(s[i]) {
			case "file":
				logger.Writers["file"] = NewFileWriter()
			case "net":
				logger.Writers["net"] = NewNetWriter()
			case "console":
				logger.Writers["console"] = NewConsoleWriter()
			}
		}
	})
	return logger
}

//获取log4go的版本号
func Version() string {
	return LOG4G_VERSION
}

//关闭日志
func (log *Logger) Close() {
	for k, w := range log.Writers {
		w.Close()
		delete(log.Writers, k)
	}
	wg.Wait()
}

//普通日志输出
func (log Logger) Log(level Level, content string) (record *Record) {
	record = log.initLogC(level, func() string {
		return content
	})
	
	return
}

//格式化日志输出
func (log Logger) LogF(level Level, format string, args ...interface{}) (record *Record) {
	record = log.initLogC(level, func() string {
		msg := format
		if len(args) > 0 {
			msg = fmt.Sprintf(format, args...)
		}
		return msg
	})
	
	return
}

//匿名函数回调输出
func (log Logger) initLogC(level Level, closure func() string) (record *Record) {
	minLevel, find := strToLevel[strings.ToUpper(log.Item("log4go.level"))]
	if find && int(level) < int(minLevel) {
		return
	}
	
	_, file, lineNum, ok := runtime.Caller(2)
	src := ""
	if ok {
		src = fmt.Sprintf("%s:%d", file, lineNum)
	}
	
	msg := closure()
	
	record = &Record{
		Level:   level,
		Content: msg,
		Size:    len(msg),
		Time:    time.Now(),
		Source:  src,
	}
	
	for _, w := range logger.Writers {
		w.Write(record)
	}
	
	return
}

//日志输出
func (log Logger) output(level Level, arg0 interface{}, args ...interface{}) (record *Record) {
	switch first := arg0.(type) {
	case string:
		if len(args) > 0 {
			record = log.LogF(level, first, args...)
		} else {
			record = log.Log(level, first)
		}
	case func() string:
		record = log.initLogC(level, first)
	default:
		format := new(bytes.Buffer)
		format.WriteString(fmt.Sprint(arg0))
		format.WriteString(strings.Repeat(" %v", len(args)))
		record = log.LogF(level, format.String(), args...)
	}
	return
}

//DEBUG 日志
func (log Logger) Debug(arg0 interface{}, args ...interface{}) {
	log.output(DEBUG, arg0, args...)
}

//INFO 日志
func (log Logger) Info(arg0 interface{}, args ...interface{}) {
	log.output(INFO, arg0, args...)
}

//WARN 日志
func (log Logger) Warn(arg0 interface{}, args ...interface{}) error {
	record := log.output(WARN, arg0, args...)
	return errors.New(record.Content)
}

//ERROR 日志
func (log Logger) Error(arg0 interface{}, args ...interface{}) error {
	record := log.output(ERROR, arg0, args...)
	return errors.New(record.Content)
}