package alisa

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-oauth2/oauth2/v4/errors"
	"github.com/pior/runnable"
	"github.com/vedga/alisa/internal/pkg/log"
	"github.com/vedga/alisa/internal/service/oauth"
)

const (
	alisaEndpointPrefix        = "/alisa"
	alisaEndpointProbe         = alisaEndpointPrefix + "/v1.0"
	alisaEndpointUserPrefix    = alisaEndpointProbe + "/user/"
	alisaEndpointUnlink        = "unlink"
	alisaEndpointDevices       = "devices"
	alisaEndpointDevicesQuery  = "query"
	alisaEndpointDevicesAction = "action"
	headerRequestID            = "X-Request-Id"
	contextUserID              = "X-User-ID"
)

// Service is Alisa service implementation
type Service struct {
	runnable.Runnable
}

// NewService return new service implementation
func NewService(router gin.IRouter, oauthService *oauth.Service) (service *Service, e error) {
	service = &Service{}

	// Following group required only authorized access
	authorized := router.Group(alisaEndpointUserPrefix)

	// Set bearer token checker
	authorized.Use(func(ginCtx *gin.Context) {
		tokenInfo, e := oauthService.ValidationBearerToken(ginCtx)
		if nil != e {
			switch e {
			case errors.ErrInvalidAccessToken:
				ginCtx.Status(http.StatusForbidden)
			default:
				ginCtx.Status(http.StatusUnauthorized)
			}

			ginCtx.Abort()
			return
		}

		// Add User ID to the context
		ginCtx.Set(contextUserID, tokenInfo.GetUserID())

		log.Log.Debugf("Token %v", tokenInfo)

		// Call next handler
		ginCtx.Next()
	})

	router.HEAD(alisaEndpointProbe, service.onProbe)
	authorized.POST(alisaEndpointUnlink, service.onUnlink)
	authorized.GET(alisaEndpointDevices, service.onDevices)
	authorized.POST(alisaEndpointDevicesQuery, service.onDevicesQuery)
	authorized.POST(alisaEndpointDevicesAction, service.onDevicesAction)

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

// onUnlink called by Yandex when accounts unlinked
func (service *Service) onUnlink(ginCtx *gin.Context) {
	log.Log.Debug("Accounts unlinked")

	msg := newUnlinkResponse(ginCtx)

	ginCtx.JSON(http.StatusOK, msg)
}

// onDevices called by Yandex to enumerate devices
func (service *Service) onDevices(ginCtx *gin.Context) {
	log.Log.Debug("Enumerate devices")

	msg := newDevicesResponse(ginCtx)

	ginCtx.JSON(http.StatusOK, msg)
}

// onDevicesQuery called by Yandex to query device states
func (service *Service) onDevicesQuery(ginCtx *gin.Context) {
	log.Log.Debug("Query devices")
	ginCtx.Status(http.StatusOK)
}

// onDevicesAction called by Yandex to perform action on the device
func (service *Service) onDevicesAction(ginCtx *gin.Context) {
	log.Log.Debug("Devices action")
	ginCtx.Status(http.StatusOK)
}
