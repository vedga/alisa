package alisa

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pior/runnable"
)

// Implementation is Alisa service implementation
type Implementation struct {
	runnable.Runnable
}

// NewService return new service implementation
func NewService() *Implementation {
	router := gin.New()
	//mux := http.NewServeMux()

	server := &http.Server{
		Addr:    "127.0.0.1:8080",
		Handler: router, //http.NotFoundHandler(),
	}

	return &Implementation{
		Runnable: runnable.HTTPServer(server),
	}
}
