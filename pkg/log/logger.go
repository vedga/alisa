package log

import (
	stdlog "log"

	"github.com/pior/runnable"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	// Log is global application logger
	Log *zap.SugaredLogger
)

// NewLogger create new loggers
func NewLogger() (logger *zap.Logger, e error) {
	if logger, e = zap.NewDevelopment(); nil != e {
		return nil, e
	}

	var stdLogger *stdlog.Logger
	if stdLogger, e = zap.NewStdLogAt(logger, zapcore.DebugLevel); nil != e {
		return logger, e
	}

	// Set logger for runnable package
	runnable.SetLogger(stdLogger)

	// Global application logger
	Log = logger.Sugar()

	return logger, nil
}
