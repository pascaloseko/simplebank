package cmd

import (
	"context"
	"database/sql"
	"log"

	"github.com/simplebank/config"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

var (
	// Used for flags.
	verbose bool

	rootCmd = &cobra.Command{
		Use:   "simple-bank",
		Short: "Simple Bank",
	}
	factories []commandFactory
)

type commandFactory func(appConfig *config.Config, tracerProvider trace.TracerProvider, propagator propagation.TextMapPropagator,
	transport *otelhttp.Transport, db *sql.DB) *cobra.Command

func addCommand(factory commandFactory) {
	factories = append(factories, factory)
}

// Execute executes the root command.
func Execute(ctx context.Context, appConfig *config.Config, tracerProvider trace.TracerProvider, propagator propagation.TextMapPropagator,
	transport *otelhttp.Transport, db *sql.DB) error {
	for _, factory := range factories {
		command := factory(appConfig, tracerProvider, propagator, transport, db)
		if command.RunE != nil {
			command.Run = printError(command.RunE)
			command.RunE = nil
		}
		rootCmd.AddCommand(command)
	}
	return rootCmd.ExecuteContext(ctx)
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
}

func printError(fn func(*cobra.Command, []string) error) func(*cobra.Command, []string) {
	return func(cmd *cobra.Command, args []string) {
		err := fn(cmd, args)
		if err != nil {
			type stackTracer interface {
				StackTrace() errors.StackTrace
			}

			var stackTracerErr stackTracer
			var ok bool
			stackTracerErr, ok = errors.Cause(err).(stackTracer)
			if !ok {
				stackTracerErr, ok = err.(stackTracer)
				if !ok {
					log.Fatal(err.Error())
				}
			}

			st := stackTracerErr.StackTrace()
			log.Fatalf("%+v", st)
		}
	}
}
