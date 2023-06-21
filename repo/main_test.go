package repo

import (
	"database/sql"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/simplebank/config"
	"github.com/simplebank/internal/testutils"
)

var dbURL string

func SetupTables(t *testing.T) (*sql.DB, func()) {
	r := require.New(t)

	db, err := sql.Open("pgx", dbURL)
	r.NoError(err)

	return db, func() {
		_, err = db.Exec("TRUNCATE \"accounts\",\"entries\",\"transfers\"")
		r.NoError(err)

		err = db.Close()
		r.NoError(err)
	}
}

func TestMain(m *testing.M) {
	appConfig, err := config.New()
	if err != nil {
		panic(err)
	}

	testutils.SeedRand()

	var finalizer func()
	dbURL, finalizer = testutils.SetupDatabase(appConfig)

	code := m.Run()

	// can't use defer since os.Exit doesn't care for defers
	finalizer()

	os.Exit(code)
}
