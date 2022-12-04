package alisa

import (
	"net/http"

	"github.com/pior/runnable"
)

// Implementation is Alisa service implementation
type Implementation struct {
	runnable.Runnable
}

// NewService return new service implementation
func NewService() *Implementation {
	server := &http.Server{
		Addr:    "127.0.0.1:8080",
		Handler: http.NotFoundHandler(),
	}

	return &Implementation{
		Runnable: runnable.HTTPServer(server),
	}
}
