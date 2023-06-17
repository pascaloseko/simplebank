package main

import (
	"context"
	"database/sql"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog/log"
	"github.com/simplebank/cmd"
	"github.com/simplebank/config"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v4/stdlib" // For pqx driver through sql

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func main() {
	var db *sql.DB
	ctx := context.Background()

	appConfig, err := config.New()
	if err != nil {
		panic(err) // TODO: replace it with zap or logrus. preferably zap
	}

	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	tracerProvider, propagator := cmd.InitOtel()
	transport := otelhttp.NewTransport(cmd.DefaultPooledTransport(), otelhttp.WithTracerProvider(tracerProvider), otelhttp.WithPropagators(propagator))
	if appConfig.Testing {
		db, err = sql.Open("pgx", appConfig.TestDBUrl)
		if err != nil {
			log.Err(err).Msg("database connection error")
			panic(err)
		}
	} else {
		db, err = sql.Open("pgx", appConfig.DBUrl)
		if err != nil {
			log.Err(err).Msg("database connection error")
			panic(err)
		}
	}
	// Establish database connection
	err = db.PingContext(ctx)
	if err != nil {
		log.Err(err).Msg("ping error")
		panic(err)
	}
	_ = cmd.Execute(ctx, appConfig, tracerProvider, propagator, transport, db)
}
