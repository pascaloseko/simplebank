package server

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"

	"github.com/simplebank/token"

	"github.com/gin-gonic/gin"

	"github.com/simplebank/config"
	"github.com/simplebank/repo"

	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

type Server struct {
	appConfig  *config.Config
	store      repo.Store
	tokenMaker token.Maker
	router     *gin.Engine
}

func NewServer(appConfig *config.Config, store repo.Store) (*Server, error) {
	tokenMaker, err := token.NewPasetoMaker(appConfig.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		err := v.RegisterValidation("currency", validCurrency)
		if err != nil {
			return nil, err
		}
	}

	server := &Server{appConfig: appConfig, store: store, tokenMaker: tokenMaker}
	return server, nil
}

// Serve serves the api endpoint
func (s *Server) Serve(ctx context.Context, _ trace.TracerProvider, _ propagation.TextMapPropagator) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	s.setupRouter()

	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", s.appConfig.Port),
		WriteTimeout: s.appConfig.WriteTimeOut,
		ReadTimeout:  s.appConfig.ReadTimeOut,
		IdleTimeout:  s.appConfig.IdleTimeOut,
		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		}}
	ch := make(chan error, 1)
	// Run our server in a goroutine so that it doesn't block.
	go func() {
		err := s.router.Run(server.Addr)
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

func (s *Server) setupRouter() {
	router := gin.Default()

	router.POST("/users", s.createUser)
	router.POST("/users/login", s.loginUser)
	router.POST("/tokens/renew_access", s.renewAccessToken)

	router.POST("/accounts", s.createAccount)
	router.GET("/accounts/:id", s.getAccount)
	router.GET("/accounts", s.listAccounts)

	router.POST("/transfers", s.createTransfer)

	s.router = router
}

func errResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
