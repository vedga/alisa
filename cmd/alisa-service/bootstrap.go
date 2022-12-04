package main

import (
	stdlog "log"

	"github.com/pior/runnable"
	"github.com/vedga/alisa/internal/service/alisa"
	"github.com/vedga/alisa/pkg/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	// Application-wide logger
	logger, e := log.NewLogger()
	if nil != e {
		stdlog.Fatal(e)
	}

	defer func(logger *zap.Logger) {
		_ = logger.Sync() // flushes buffer, if any
	}(logger)

	// Set logger for runnable package
	var stdLogger *stdlog.Logger
	if stdLogger, e = zap.NewStdLogAt(logger, zapcore.DebugLevel); nil != e {
		stdlog.Fatal(e)
	}
	runnable.SetLogger(stdLogger)

	appManager := runnable.NewManager()

	// Create Alisa service and add it to the application manager
	alisaService := alisa.NewService()
	appManager.Add(alisaService)

	log.Log.Debugf("Application started")

	runnable.Run(appManager.Build())

	log.Log.Debugf("Shutdown application operation complete")
}
