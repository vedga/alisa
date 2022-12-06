package main

import (
	stdlog "log"

	"github.com/gin-gonic/gin"
	"github.com/pior/runnable"
	"github.com/vedga/alisa/internal/pkg/log"
	"github.com/vedga/alisa/internal/service/alisa"
	"github.com/vedga/alisa/internal/service/httpserver"
	"github.com/vedga/alisa/internal/service/mqtt"
	"github.com/vedga/alisa/internal/service/oauth"
	"github.com/vedga/alisa/pkg/eventbus"
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

	bus := eventbus.New()

	appManager := runnable.NewManager()

	var mqttService *mqtt.Service
	if mqttService, e = mqtt.NewService(bus); nil != e {
		stdlog.Fatal(e)
	}

	var httpService *httpserver.Service
	if httpService, e = httpserver.NewService(); nil != e {
		stdlog.Fatal(e)
	}

	var oauthService *oauth.Service
	if oauthService, e = oauth.NewService(httpService.Router()); nil != e {
		stdlog.Fatal(e)
	}

	// Create Alisa service and add it to the application manager
	var alisaService *alisa.Service
	if alisaService, e = alisa.NewService(httpService.Router(), oauthService); nil != e {
		stdlog.Fatal(e)
	}

	appManager.Add(httpService, oauthService, alisaService, mqttService)

	log.Log.Debugf("Application started")

	runnable.Run(appManager.Build())

	log.Log.Debugf("Shutdown application operation complete")
}
