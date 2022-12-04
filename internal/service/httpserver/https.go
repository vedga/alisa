package httpserver

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/pior/runnable"
)

type httpServerTLS struct {
	server          *http.Server
	shutdownTimeout time.Duration
}

// serverTLS returns a runnable that runs a *http.Server.
func serverTLS(server *http.Server) runnable.Runnable {
	return &httpServerTLS{server, time.Second * 30}
}

// Run is implementation of Runnable interface
func (r *httpServerTLS) Run(ctx context.Context) error {
	errChan := make(chan error)

	go func() {
		log.Printf("http_server: listening on %s", r.server.Addr)
		errChan <- r.server.ListenAndServeTLS("", "")
	}()

	var err error
	var shutdownErr error

	select {
	case <-ctx.Done():
		log.Printf("http_server: shutdown")
		shutdownErr = r.shutdown()
		err = <-errChan
	case err = <-errChan:
		log.Printf("http_server: shutdown (err: %s)", err)
		shutdownErr = r.shutdown()
	}

	if err == http.ErrServerClosed {
		err = nil
	}
	if err == nil && shutdownErr != nil {
		err = fmt.Errorf("server shutdown: %w", shutdownErr)
	}

	return err
}

func (r *httpServerTLS) shutdown() error {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, r.shutdownTimeout)
	defer cancel()

	return r.server.Shutdown(ctx)
}
