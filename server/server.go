package server

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"net/http"

	"github.com/simplebank/config"
	"github.com/simplebank/repo"

	"github.com/gorilla/mux"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

type Server struct {
	appConfig *config.Config
	queries   *repo.Queries
}

func NewServer(appConfig *config.Config, db *sql.DB) *Server {
	queries := repo.New(db)
	return &Server{appConfig: appConfig, queries: queries}
}

// Serve serves the api endpoint
func (s *Server) Serve(ctx context.Context, tracerProvider trace.TracerProvider, propagator propagation.TextMapPropagator) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	router := mux.NewRouter()
	router.Use(otelmux.Middleware("server", otelmux.WithTracerProvider(tracerProvider), otelmux.WithPropagators(propagator)))

	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", s.appConfig.Port),
		WriteTimeout: s.appConfig.WriteTimeOut,
		ReadTimeout:  s.appConfig.ReadTimeOut,
		IdleTimeout:  s.appConfig.IdleTimeOut,
		Handler:      router,
		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		}}
	ch := make(chan error, 1)
	// Run our server in a goroutine so that it doesn't block.
	go func() {
		err := server.ListenAndServe()
		ch <- err
		close(ch)
	}()
	select {
	case err := <-ch:
		return err
	case <-ctx.Done():
		err := server.Shutdown(ctx)
		return err
	}
}
