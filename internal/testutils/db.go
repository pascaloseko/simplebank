package testutils

import (
	"context"
	"math/rand"
	"net/url"
	"strconv"
	"strings"

	"github.com/simplebank/config"

	"github.com/golang-migrate/migrate/v4"
	"github.com/jackc/pgx/v4"
)

const (
	migrationsFilepath = "file:///app/db/migrations"
)

func SetupDatabase(appConfig *config.Config) (string, func()) {
	dsn, err := url.Parse(appConfig.DBUrl)
	if err != nil {
		panic(err)
	}

	SeedRand()

	randomSuffix := strconv.Itoa(rand.Intn(1000))
	dbName := "simplebank_" + randomSuffix

	oldPath := dsn.Path
	dsn.Path = ""
	globalConn, err := pgx.Connect(context.Background(), dsn.String())
	if err != nil {
		panic(err)
	}

	_, err = globalConn.Exec(context.Background(), "CREATE DATABASE "+dbName)
	if err != nil {
		panic(err)
	}

	dsn.Path = strings.Replace(oldPath, "user", dbName, 1)

	migrator, err := migrate.New(migrationsFilepath, dsn.String())
	if err != nil {
		panic(err)
	}
	err = migrator.Up()
	if err != nil {
		panic(err)
	}
	return dsn.String(), func() {
		err = migrator.Down()
		if err != nil {
			panic(err)
		}

		_, err = migrator.Close()
		if err != nil {
			panic(err)
		}

		_, err = globalConn.Exec(context.Background(), "DROP DATABASE "+dbName)
		if err != nil {
			panic(err)
		}
	}
}
