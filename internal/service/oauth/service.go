package oauth

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-oauth2/oauth2/v4"
	"github.com/go-oauth2/oauth2/v4/errors"
	"github.com/go-oauth2/oauth2/v4/manage"
	"github.com/go-oauth2/oauth2/v4/models"
	oauthserver "github.com/go-oauth2/oauth2/v4/server"
	"github.com/go-oauth2/oauth2/v4/store"
	"github.com/pior/runnable"
)

const (
	envYandexClientID      = "YANDEX_CLIENT_ID"
	envYandexClientSecret  = "YANDEX_CLIENT_SECRET"
	envCallbackURL         = "CALLBACK_URL"
	yandexCallbackValue    = "https://social.yandex.net/broker/redirect"
	oauthEndpointPrefix    = "/oauth"
	oauthEndpointAuthorize = oauthEndpointPrefix + "/authorize"
	oauthEndpointToken     = oauthEndpointPrefix + "/token"
)

// Service is Alisa service implementation
type Service struct {
	runnable.Runnable
	oauthServer *oauthserver.Server
}

// NewService return new service implementation
func NewService(router gin.IRoutes) (service *Service, e error) {
	service = &Service{}

	manager := manage.NewDefaultManager()
	// token memory store
	manager.MustTokenStorage(store.NewMemoryTokenStore())

	clientID := ""
	if value, found := os.LookupEnv(envYandexClientID); found {
		clientID = value
	}

	clientSecret := ""
	if value, found := os.LookupEnv(envYandexClientSecret); found {
		clientSecret = value
	}

	callback := yandexCallbackValue
	if value, found := os.LookupEnv(envCallbackURL); found {
		callback = value
	}

	// client memory store
	clientStore := store.NewClientStore()
	_ = clientStore.Set(clientID, &models.Client{
		ID:     "000000",
		Secret: clientSecret,
		Domain: callback,
	})
	manager.MapClientStorage(clientStore)

	service.oauthServer = oauthserver.NewDefaultServer(manager)
	service.oauthServer.SetAllowGetAccessRequest(true)
	service.oauthServer.SetClientInfoHandler(oauthserver.ClientFormHandler)

	service.oauthServer.UserAuthorizationHandler = func(w http.ResponseWriter, r *http.Request) (string, error) {
		if clientID := strings.TrimSpace(r.FormValue("client_id")); len(clientID) > 0 {
			return clientID, nil
		}

		return "", errors.ErrAccessDenied
	}

	service.oauthServer.SetInternalErrorHandler(func(err error) (re *errors.Response) {
		log.Println("Internal Error:", err.Error())
		return
	})

	service.oauthServer.SetResponseErrorHandler(func(re *errors.Response) {
		log.Println("Response Error:", re.Error.Error())
	})

	router.GET(oauthEndpointAuthorize, service.onAuthorize)
	router.POST(oauthEndpointAuthorize, service.onAuthorize)
	router.GET(oauthEndpointToken, service.onToken)
	router.POST(oauthEndpointToken, service.onToken)

	return service, nil
}

// Run is implementation of runnable.Runnable interface
func (service *Service) Run(ctx context.Context) error {
	// Wait until operation complete
	<-ctx.Done()

	return ctx.Err()
}

// onAuthorize implement user authorization
func (service *Service) onAuthorize(ginCtx *gin.Context) {
	_ = service.oauthServer.HandleAuthorizeRequest(ginCtx.Writer, ginCtx.Request)
}

// onToken implement token issuing
func (service *Service) onToken(ginCtx *gin.Context) {
	_ = service.oauthServer.HandleTokenRequest(ginCtx.Writer, ginCtx.Request)
}

// ValidationBearerToken do validate token on Resource Service
func (service *Service) ValidationBearerToken(ginCtx *gin.Context) (oauth2.TokenInfo, error) {
	return service.oauthServer.ValidationBearerToken(ginCtx.Request)
}
