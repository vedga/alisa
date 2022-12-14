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

// Service is HTTP(S) service implementation
type Service struct {
	runnable.Runnable
	engine *gin.Engine
}

// NewService return new service implementation
func NewService() (service *Service, e error) {
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

	service = &Service{
		engine: gin.New(),
	}

	server := &http.Server{
		Addr:      "0.0.0.0:8443",
		TLSConfig: tlsConfig,
		Handler:   service.engine,
	}

	if nil == tlsConfig {
		service.Runnable = runnable.HTTPServer(server)
	} else {
		service.Runnable = serverTLS(server)
	}

	return service, nil
}

// Router return routes controller
func (service *Service) Router() gin.IRouter {
	return service.engine
}
