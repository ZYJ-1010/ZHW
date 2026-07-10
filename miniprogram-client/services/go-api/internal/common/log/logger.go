package log

import stdlog "log"

func Infof(format string, args ...interface{}) {
	stdlog.Printf(format, args...)
}
