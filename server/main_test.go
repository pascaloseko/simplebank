package server

import (
	"os"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/simplebank/config"
	"github.com/simplebank/repo"
)

func newTestServer(t *testing.T, store repo.Store) *Server {
	appConfig, err := config.New()
	if err != nil {
		t.Fatalf("failed to create app config: %s", err)
	}

	server := NewServer(appConfig, store)

	return server
}

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)

	os.Exit(m.Run())
}
