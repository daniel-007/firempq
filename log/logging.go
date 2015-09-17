package log

import "os"
import "github.com/op/go-logging"

func InitLogging(level logging.Level) {
	format := logging.MustStringFormatter(
		"%{color}%{time:2006-01-02 15:04:05.00000}: %{level}%{color:reset} %{shortfile} %{message}",
	)
	logbackend := logging.NewLogBackend(os.Stderr, "", 0)
	formatter := logging.NewBackendFormatter(logbackend, format)
	logging.SetBackend(formatter)
	logging.SetLevel(level, "firempq")
	fixLogger()
}

func fixLogger() {
	Logger.ExtraCalldepth = 1
}

var Logger = logging.MustGetLogger("firempq")

var Error func(string, ...interface{}) = Logger.Error
var Critical func(string, ...interface{}) = Logger.Critical
var Warning func(string, ...interface{}) = Logger.Warning
var Notice func(string, ...interface{}) = Logger.Notice
var Info func(string, ...interface{}) = Logger.Info
var Debug func(string, ...interface{}) = Logger.Debug