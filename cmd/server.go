package cmd

import (
	"context"
	"database/sql"
	"net/http"
	"os"

	"github.com/pkg/errors"
	"github.com/simplebank/repo"

	"github.com/simplebank/config"
	"github.com/simplebank/server"

	"github.com/spf13/cobra"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

func init() {
	addCommand(serverCmdFactory)
}

func serverCmdFactory(appConfig *config.Config, tracerProvider trace.TracerProvider, propagator propagation.TextMapPropagator,
	_ *otelhttp.Transport, db *sql.DB) *cobra.Command {
	return &cobra.Command{
		Use:   "server",
		Short: "Run simplebank api server",
		RunE: func(cmd *cobra.Command, args []string) error {
			if appConfig.RuntimeEnvironment == "cloud" && appConfig.Appenv == "backops" {
				appConfig.Appenv = os.Getenv("GOOGLE_CLOUD_PROJECT")
			}
			store := repo.NewStore(db)
			api := server.NewServer(appConfig, store)
			err := api.Serve(cmd.Context(), tracerProvider, propagator)
			if err != nil && !errors.Is(err, context.Canceled) && !errors.Is(err, http.ErrServerClosed) {
				return err
			}
			return nil
		},
	}
}
