package httpserver

import (
	"crypto/tls"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/pior/runnable"
)

const (
	envCertificateChain = "CERTIFICATE_CHAIN"
	envPrivateKey       = "PRIVATE_KEY"
)

// Implementation is Alisa service implementation
type Implementation struct {
	runnable.Runnable
	engine *gin.Engine
}

// NewService return new service implementation
func NewService() (implementation *Implementation, e error) {
	var tlsConfig *tls.Config
	if fileName, found := os.LookupEnv(envCertificateChain); found {
		// TLS configuration
		var certificateChain []byte
		if certificateChain, e = os.ReadFile(fileName); nil != e {
			return nil, e
		}

		var privateKey []byte
		if fileName, found = os.LookupEnv(envPrivateKey); found {
			if privateKey, e = os.ReadFile(fileName); nil != e {
				return nil, e
			}
		}

		var certificate tls.Certificate
		if certificate, e = tls.X509KeyPair(certificateChain, privateKey); nil != e {
			return nil, e
		}

		tlsConfig = &tls.Config{
			Certificates: []tls.Certificate{
				certificate,
			},
		}
	}

	implementation = &Implementation{
		engine: gin.New(),
	}

	server := &http.Server{
		Addr:      "0.0.0.0:8443",
		TLSConfig: tlsConfig,
		Handler:   implementation.engine,
	}

	if nil == tlsConfig {
		implementation.Runnable = runnable.HTTPServer(server)
	} else {
		implementation.Runnable = HTTPServerTLS(server)
	}

	return implementation, nil
}

// Routes return routes controller
func (implementation *Implementation) Routes() gin.IRoutes {
	return implementation.engine
}
