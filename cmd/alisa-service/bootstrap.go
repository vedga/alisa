package main

import (
	stdlog "log"

	"go.uber.org/zap"

	"github.com/pior/runnable"
	"github.com/vedga/alisa/internal/service/alisa"
	"github.com/vedga/alisa/pkg/log"
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

	appManager := runnable.NewManager()

	// Create Alisa service and add it to the application manager
	alisaService := alisa.NewService()
	appManager.Add(alisaService)

	log.Log.Debugf("Application started")

	runnable.Run(appManager.Build())

	log.Log.Debugf("Shutdown application operation complete")
}
