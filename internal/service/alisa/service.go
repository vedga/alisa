package alisa

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pior/runnable"
	"github.com/vedga/alisa/internal/pkg/log"
)

const (
	backendEndpointPrefix = "/alisa/"
	backendEndpointProbe  = backendEndpointPrefix + "v1.0/"
)

// Implementation is Alisa service implementation
type Implementation struct {
	runnable.Runnable
}

// NewService return new service implementation
func NewService(router gin.IRoutes) (implementation *Implementation, e error) {
	implementation = &Implementation{}

	router.HEAD(backendEndpointProbe, implementation.onProbe)
	// Debug only!
	router.GET(backendEndpointProbe, implementation.onProbe)

	return implementation, nil
}

// Run is implementation of runnable.Runnable interface
func (implementation *Implementation) Run(ctx context.Context) error {
	// Wait until operation complete
	<-ctx.Done()

	return ctx.Err()
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
