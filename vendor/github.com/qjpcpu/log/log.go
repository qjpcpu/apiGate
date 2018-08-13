package log

import (
	"github.com/qjpcpu/log/logging"
	syslog "log"
	"os"
	"path/filepath"
	"strings"
)

// package global variables
var lg *logging.Logger
var setLogLevel func(Level)
var log_option = defaultLogOption()

const (
	NormFormat = "%{level}: [%{time:2006-01-02 15:04:05.000}][gid:%{goroutineid}/gcnt:%{goroutinecount}][%{shortfile}][%{message}]"
)

type Level int

const (
	CRITICAL Level = iota + 1
	ERROR
	WARNING
	NOTICE
	INFO
	DEBUG
)

func (lvl Level) loggingLevel() logging.Level {
	return logging.Level(lvl - 1)
}

func ParseLogLevel(lstr string) Level {
	lstr = strings.ToLower(lstr)
	switch lstr {
	case "critical":
		return CRITICAL
	case "error":
		return ERROR
	case "warning":
		return WARNING
	case "notice":
		return NOTICE
	case "info":
		return INFO
	case "debug":
		return DEBUG
	default:
		return INFO
	}
}

type LogOption struct {
	LogFile    string
	Level      Level
	Format     string
	RotateType logging.RotateType
	files      []*logging.FileLogWriter
}

func defaultLogOption() LogOption {
	return LogOption{
		Level:      DEBUG,
		Format:     NormFormat,
		RotateType: logging.RotateDaily,
	}
}

func init() {
	InitLog(defaultLogOption())
}

func InitLog(opt LogOption) {
	if len(log_option.files) > 0 {
		for _, f := range log_option.files {
			if f != nil {
				f.Close()
			}
		}
		log_option.files = nil
	}
	if opt.Format == "" {
		opt.Format = NormFormat
	}
	if opt.Level <= 0 {
		opt.Level = INFO
	}
	format := logging.MustStringFormatter(opt.Format)
	if opt.LogFile != "" {
		// mkdir log dir
		os.MkdirAll(filepath.Dir(opt.LogFile), 0777)
		filename := opt.LogFile
		info_log_fp, err := logging.NewFileLogWriter(filename, opt.RotateType)
		if err != nil {
			syslog.Fatalf("open file[%s] failed[%s]", filename, err)
		}

		err_log_fp, err := logging.NewFileLogWriter(filename+".wf", opt.RotateType)
		if err != nil {
			syslog.Fatalf("open file[%s.wf] failed[%s]", filename, err)
		}
		opt.files = []*logging.FileLogWriter{info_log_fp, err_log_fp}

		backend_info := logging.NewLogBackend(info_log_fp, "", 0)
		backend_err := logging.NewLogBackend(err_log_fp, "", 0)
		backend_info_formatter := logging.NewBackendFormatter(backend_info, format)
		backend_err_formatter := logging.NewBackendFormatter(backend_err, format)

		backend_info_leveld := logging.AddModuleLevel(backend_info_formatter)
		backend_info_leveld.SetLevel(opt.Level.loggingLevel(), "")

		backend_err_leveld := logging.AddModuleLevel(backend_err_formatter)
		backend_err_leveld.SetLevel(logging.ERROR, "")
		logging.SetBackend(backend_info_leveld, backend_err_leveld)

		// set log level handler
		setLogLevel = func(lvl Level) {
			backend_info_leveld.SetLevel(lvl.loggingLevel(), "")
		}
	} else {
		backend1 := logging.NewLogBackend(os.Stderr, "", 0)
		backend1Formatter := logging.NewBackendFormatter(backend1, format)
		backend1Leveled := logging.AddModuleLevel(backend1Formatter)
		backend1Leveled.SetLevel(opt.Level.loggingLevel(), "")
		logging.SetBackend(backend1Leveled)
		// set log level handler
		setLogLevel = func(lvl Level) {
			backend1Leveled.SetLevel(lvl.loggingLevel(), "")
		}
	}
	lg = logging.MustGetLogger("")
	lg.ExtraCalldepth += 1
	log_option = opt
}

func isformat(format string) bool {
	end := len(format)
	unfound := -2
	var istag int = unfound
	for i := 0; i < end; i++ {
		if format[i] == '%' {
			if istag == i-1 {
				istag = unfound
			} else {
				istag = i
			}
		} else {
			if istag == i-1 {
				return true
			}
		}
	}
	return false
}

func Infof(format string, args ...interface{}) {
	if lg == nil {
		return
	}
	lg.Infof(format, args...)
}

func Warningf(format string, args ...interface{}) {
	if lg == nil {
		return
	}
	lg.Warningf(format, args...)
}

func Criticalf(format string, args ...interface{}) {
	if lg == nil {
		return
	}
	lg.Criticalf(format, args...)
}

func Errorf(format string, args ...interface{}) {
	if lg == nil {
		return
	}
	lg.Errorf(format, args...)
}

func Debugf(format string, args ...interface{}) {
	if lg == nil {
		return
	}
	lg.Debugf(format, args...)
}

func Noticef(format string, args ...interface{}) {
	if lg == nil {
		return
	}
	lg.Noticef(format, args...)
}

func isformatLog(args ...interface{}) bool {
	for loop := true; loop; loop = false {
		if len(args) <= 1 {
			break
		}
		format, ok := args[0].(string)
		if !ok {
			break
		}
		return isformat(format)
	}
	return false
}

func Info(args ...interface{}) {
	if lg == nil {
		return
	}
	if isformatLog(args...) {
		lg.Infof(args[0].(string), args[1:]...)
	} else {
		lg.Infof(strings.TrimSpace(strings.Repeat("%+v ", len(args))), args...)
	}
}

func Warning(args ...interface{}) {
	if lg == nil {
		return
	}
	if isformatLog(args...) {
		lg.Warningf(args[0].(string), args[1:]...)
	} else {
		lg.Warningf(strings.TrimSpace(strings.Repeat("%+v ", len(args))), args...)
	}
}

func Critical(args ...interface{}) {
	if lg == nil {
		return
	}
	if isformatLog(args...) {
		lg.Criticalf(args[0].(string), args[1:]...)
	} else {
		lg.Criticalf(strings.TrimSpace(strings.Repeat("%+v ", len(args))), args...)
	}
}

func Error(args ...interface{}) {
	if lg == nil {
		return
	}
	if isformatLog(args...) {
		lg.Errorf(args[0].(string), args[1:]...)
	} else {
		lg.Errorf(strings.TrimSpace(strings.Repeat("%+v ", len(args))), args...)
	}
}

func Debug(args ...interface{}) {
	if lg == nil {
		return
	}
	if isformatLog(args...) {
		lg.Debugf(args[0].(string), args[1:]...)
	} else {
		lg.Debugf(strings.TrimSpace(strings.Repeat("%+v ", len(args))), args...)
	}
}

func Notice(args ...interface{}) {
	if lg == nil {
		return
	}
	if isformatLog(args...) {
		lg.Noticef(args[0].(string), args[1:]...)
	} else {
		lg.Noticef(strings.TrimSpace(strings.Repeat("%+v ", len(args))), args...)
	}
}

func SetLogLevel(lvl Level) {
	if setLogLevel != nil {
		setLogLevel(lvl)
		log_option.Level = lvl
	}
}

func GetLogLevel() Level {
	return log_option.Level
}
