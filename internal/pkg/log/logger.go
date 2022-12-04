package log

import (
	"go.uber.org/zap"
)

var (
	// Log is global application logger
	Log *zap.SugaredLogger
)

// NewLogger create new loggers. Also set
func NewLogger() (logger *zap.Logger, e error) {
	if logger, e = zap.NewDevelopment(); nil != e {
		return nil, e
	}

	// Global application logger
	Log = logger.Sugar()

	return logger, nil
}
