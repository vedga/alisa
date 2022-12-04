package main

import (
	stdlog "log"

	"github.com/gin-gonic/gin"
	"github.com/pior/runnable"
	"github.com/vedga/alisa/internal/pkg/log"
	"github.com/vedga/alisa/internal/service/alisa"
	"github.com/vedga/alisa/internal/service/httpserver"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zapio"
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

	// Set Runnable framework logger
	var stdLogger *stdlog.Logger
	if stdLogger, e = zap.NewStdLogAt(logger, zapcore.DebugLevel); nil != e {
		stdlog.Fatal(e)
	}
	runnable.SetLogger(stdLogger)

	// Set gin framework writer
	logWriter := &zapio.Writer{
		Log:   logger,
		Level: zap.DebugLevel,
	}
	defer func() {
		_ = logWriter.Close()
	}()
	gin.DefaultWriter = logWriter
	gin.DefaultErrorWriter = logWriter

	appManager := runnable.NewManager()

	var httpService *httpserver.Implementation
	if httpService, e = httpserver.NewService(); nil != e {
		stdlog.Fatal(e)
	}

	// Create Alisa service and add it to the application manager
	var alisaService *alisa.Implementation
	if alisaService, e = alisa.NewService(httpService.Routes()); nil != e {
		stdlog.Fatal(e)
	}
	appManager.Add(alisaService, httpService)

	log.Log.Debugf("Application started")

	runnable.Run(appManager.Build())

	log.Log.Debugf("Shutdown application operation complete")
}
