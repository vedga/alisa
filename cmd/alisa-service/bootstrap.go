package main

import (
	"crypto/tls"
	stdlog "log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/pior/runnable"
	"github.com/vedga/alisa/internal/pkg/log"
	"github.com/vedga/alisa/internal/service/alisa"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zapio"
)

const (
	envCertificateChain = "CERTIFICATE_CHAIN"
	envPrivateKey       = "PRIVATE_KEY"
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

	var tlsConfig *tls.Config
	if fileName, found := os.LookupEnv(envCertificateChain); found {
		// TLS configuration
		var certificateChain []byte
		if certificateChain, e = os.ReadFile(fileName); nil != e {
			log.Log.Error("Unable to read certificate file", zap.Error(e))
			logger.Error("Unable to read certificate file", zap.Error(e))
			return
		}

		var privateKey []byte
		if fileName, found = os.LookupEnv(envPrivateKey); found {
			if privateKey, e = os.ReadFile(fileName); nil != e {
				log.Log.Error("Unable to read private key file", zap.Error(e))
			}
		}

		var certificate tls.Certificate
		if certificate, e = tls.X509KeyPair(certificateChain, privateKey); nil != e {
			stdlog.Fatal(e)
		}

		tlsConfig = &tls.Config{
			Certificates: []tls.Certificate{
				certificate,
			},
		}
	}

	// Create Alisa service and add it to the application manager
	var alisaService *alisa.Implementation
	if alisaService, e = alisa.NewService(tlsConfig); nil != e {
		stdlog.Fatal(e)
	}
	appManager.Add(alisaService)

	log.Log.Debugf("Application started")

	runnable.Run(appManager.Build())

	log.Log.Debugf("Shutdown application operation complete")
}
