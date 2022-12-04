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

// Service is Alisa service implementation
type Service struct {
	runnable.Runnable
}

// NewService return new service implementation
func NewService(router gin.IRoutes) (service *Service, e error) {
	service = &Service{}

	router.HEAD(backendEndpointProbe, service.onProbe)
	// Debug only!
	router.GET(backendEndpointProbe, service.onProbe)

	return service, nil
}

// Run is implementation of runnable.Runnable interface
func (service *Service) Run(ctx context.Context) error {
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
func (service *Service) onProbe(ginCtx *gin.Context) {
	log.Log.Debug("Service probed")
	ginCtx.Status(http.StatusOK)
}
