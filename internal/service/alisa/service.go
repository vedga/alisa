package alisa

import (
	"crypto/tls"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pior/runnable"
	"github.com/vedga/alisa/pkg/log"
)

const (
	endpointProbe = "/v1.0/"
)

// Implementation is Alisa service implementation
type Implementation struct {
	runnable.Runnable
}

// NewService return new service implementation
func NewService(tlsConfig *tls.Config) (implementation *Implementation, e error) {
	implementation = &Implementation{}

	router := gin.New()

	router.HEAD(endpointProbe, implementation.onProbe)
	// Debug only!
	router.GET(endpointProbe, implementation.onProbe)

	server := &http.Server{
		Addr:      "0.0.0.0:8443",
		TLSConfig: tlsConfig,
		Handler:   router,
	}

	if nil == tlsConfig {
		implementation.Runnable = runnable.HTTPServer(server)
	} else {
		implementation.Runnable = HTTPServerTLS(server)
	}

	return implementation, nil
}

// onProbe called by Yandex to check this service ready status
// Possible code responses:
// http.StatusOK - service ready
// http.StatusBadRequest - request error
// http.StatusNotFound - URL not found
// StatusInternalServerError - internal service error
func (implementation *Implementation) onProbe(ginCtx *gin.Context) {
	log.Log.Debug("Service probed")
	ginCtx.Status(http.StatusOK)
}
