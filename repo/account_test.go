package repo

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v4/stdlib" // For pqx driver through sql
	_ "github.com/lib/pq"
	"github.com/simplebank/config"
	"github.com/simplebank/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func createRandomAccount(t *testing.T) Account {
	ctx := context.Background()

	db, finalizer := SetupTables(t)
	t.Cleanup(finalizer)

	r := New(db)

	arg := CreateAccountParams{
		Owner:    testutils.RandomOwner(),
		Balance:  testutils.RandomMoney(),
		Currency: testutils.RandomCurrency(),
	}

	account, err := r.CreateAccount(ctx, arg)
	assert.NoError(t, err)
	assert.NotEmpty(t, account)
	assert.Equal(t, arg.Owner, account.Owner)
	assert.Equal(t, arg.Balance, account.Balance)
	assert.Equal(t, arg.Currency, account.Currency)

	assert.NotZero(t, account.ID)
	assert.NotZero(t, account.CreatedAt)

	return account
}

func TestCreateAccount(t *testing.T) {
	createRandomAccount(t)
}

func TestGetAccount(t *testing.T) {
	acc1 := createRandomAccount(t)
	ctx := context.Background()

	db, finalizer := SetupTables(t)
	t.Cleanup(finalizer)

	r := New(db)

	acc2, err := r.GetAccount(ctx, acc1.ID)
	assert.NoError(t, err)
	assert.NotEmpty(t, acc2)

	assert.Equal(t, acc1.ID, acc2.ID)
	assert.Equal(t, acc1.Owner, acc2.Owner)
	assert.Equal(t, acc1.Balance, acc2.Balance)
	assert.Equal(t, acc1.Currency, acc2.Currency)
	assert.WithinDuration(t, acc1.CreatedAt, acc2.CreatedAt, time.Second)
}

func TestUpdateAccount(t *testing.T) {
	ctx := context.Background()

	db, finalizer := SetupTables(t)
	t.Cleanup(finalizer)

	r := New(db)

	account := createRandomAccount(t)

	arg := UpdateAccountParams{
		ID:      account.ID,
		Balance: testutils.RandomMoney(),
	}

	account2, err := r.UpdateAccount(ctx, arg)
	assert.NoError(t, err)
	assert.NotEmpty(t, account2)

	assert.Equal(t, account.ID, account2.ID)
	assert.Equal(t, account.Owner, account2.Owner)
	assert.Equal(t, arg.Balance, account2.Balance)
	assert.Equal(t, account.Currency, account2.Currency)
	assert.WithinDuration(t, account.CreatedAt, account2.CreatedAt, time.Second)
}

func TestDeleteAccount(t *testing.T) {
	ctx := context.Background()

	db, finalizer := SetupTables(t)
	t.Cleanup(finalizer)

	r := New(db)

	account1 := createRandomAccount(t)
	err := r.DeleteAccount(ctx, account1.ID)
	assert.NoError(t, err)

	account2, err := r.GetAccount(ctx, account1.ID)
	assert.EqualError(t, err, sql.ErrNoRows.Error())
	assert.Empty(t, account2)
}

func TestListAccounts(t *testing.T) {
	ctx := context.Background()

	db, finalizer := SetupTables(t)
	t.Cleanup(finalizer)

	r := New(db)

	var lastAccount Account
	for i := 0; i < 10; i++ {
		lastAccount = createRandomAccount(t)
	}

	arg := ListAccountsParams{
		Owner:  lastAccount.Owner,
		Limit:  5,
		Offset: 0,
	}

	accounts, err := r.ListAccounts(ctx, arg)
	require.NoError(t, err)
	require.NotEmpty(t, accounts)

	for _, account := range accounts {
		require.NotEmpty(t, account)
		require.Equal(t, lastAccount.Owner, account.Owner)
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
