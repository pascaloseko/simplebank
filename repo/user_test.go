package repo

import (
	"context"
	"testing"
	"time"

	"github.com/simplebank/internal/testutils"
	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v4"
)

func createRandomUser(t *testing.T) User {
	hashedPassword, err := testutils.HashPassword(testutils.RandomString(6))
	require.NoError(t, err)

	db, finalizer := SetupTables(t)
	t.Cleanup(finalizer)

	r := New(db)

	arg := CreateUserParams{
		Username:       testutils.RandomOwner(),
		HashedPassword: hashedPassword,
		FullName:       testutils.RandomOwner(),
		Email:          testutils.RandomEmail(),
	}

	user, err := r.CreateUser(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, user)

	require.Equal(t, arg.Username, user.Username)
	require.Equal(t, arg.HashedPassword, user.HashedPassword)
	require.Equal(t, arg.FullName, user.FullName)
	require.Equal(t, arg.Email, user.Email)
	require.True(t, user.PasswordChangedAt.IsZero())
	require.NotZero(t, user.CreatedAt)

	return user
}

func TestCreateUser(t *testing.T) {
	createRandomUser(t)
}

func TestGetUser(t *testing.T) {
	db, finalizer := SetupTables(t)
	t.Cleanup(finalizer)

	r := New(db)

	user1 := createRandomUser(t)
	user2, err := r.GetUser(context.Background(), user1.Username)
	require.NoError(t, err)
	require.NotEmpty(t, user2)

	require.Equal(t, user1.Username, user2.Username)
	require.Equal(t, user1.HashedPassword, user2.HashedPassword)
	require.Equal(t, user1.FullName, user2.FullName)
	require.Equal(t, user1.Email, user2.Email)
	require.WithinDuration(t, user1.PasswordChangedAt, user2.PasswordChangedAt, time.Second)
	require.WithinDuration(t, user1.CreatedAt, user2.CreatedAt, time.Second)
}

func TestUpdateUserOnlyFullName(t *testing.T) {
	db, finalizer := SetupTables(t)
	t.Cleanup(finalizer)

	r := New(db)

	oldUser := createRandomUser(t)

	newFullName := testutils.RandomOwner()
	updatedUser, err := r.UpdateUser(context.Background(), UpdateUserParams{
		Username: oldUser.Username,
		FullName: null.StringFrom(newFullName),
	})

	require.NoError(t, err)
	require.NotEqual(t, oldUser.FullName, updatedUser.FullName)
	require.Equal(t, newFullName, updatedUser.FullName)
	require.Equal(t, oldUser.Email, updatedUser.Email)
	require.Equal(t, oldUser.HashedPassword, updatedUser.HashedPassword)
}

func TestUpdateUserOnlyEmail(t *testing.T) {
	db, finalizer := SetupTables(t)
	t.Cleanup(finalizer)

	r := New(db)

	oldUser := createRandomUser(t)

	newEmail := testutils.RandomEmail()
	updatedUser, err := r.UpdateUser(context.Background(), UpdateUserParams{
		Username: oldUser.Username,
		Email:    null.StringFrom(newEmail),
	})

	require.NoError(t, err)
	require.NotEqual(t, oldUser.Email, updatedUser.Email)
	require.Equal(t, newEmail, updatedUser.Email)
	require.Equal(t, oldUser.FullName, updatedUser.FullName)
	require.Equal(t, oldUser.HashedPassword, updatedUser.HashedPassword)
}

func TestUpdateUserOnlyPassword(t *testing.T) {
	db, finalizer := SetupTables(t)
	t.Cleanup(finalizer)

	r := New(db)

	oldUser := createRandomUser(t)

	newPassword := testutils.RandomString(6)
	newHashedPassword, err := testutils.HashPassword(newPassword)
	require.NoError(t, err)

	updatedUser, err := r.UpdateUser(context.Background(), UpdateUserParams{
		Username:       oldUser.Username,
		HashedPassword: null.StringFrom(newHashedPassword),
	})

	require.NoError(t, err)
	require.NotEqual(t, oldUser.HashedPassword, updatedUser.HashedPassword)
	require.Equal(t, newHashedPassword, updatedUser.HashedPassword)
	require.Equal(t, oldUser.FullName, updatedUser.FullName)
	require.Equal(t, oldUser.Email, updatedUser.Email)
}

func TestUpdateUserAllFields(t *testing.T) {
	db, finalizer := SetupTables(t)
	t.Cleanup(finalizer)

	r := New(db)

	oldUser := createRandomUser(t)

	newFullName := testutils.RandomOwner()
	newEmail := testutils.RandomEmail()
	newPassword := testutils.RandomString(6)
	newHashedPassword, err := testutils.HashPassword(newPassword)
	require.NoError(t, err)

	updatedUser, err := r.UpdateUser(context.Background(), UpdateUserParams{
		Username:       oldUser.Username,
		FullName:       null.StringFrom(newFullName),
		Email:          null.StringFrom(newEmail),
		HashedPassword: null.StringFrom(newHashedPassword),
	})

	require.NoError(t, err)
	require.NotEqual(t, oldUser.HashedPassword, updatedUser.HashedPassword)
	require.Equal(t, newHashedPassword, updatedUser.HashedPassword)
	require.NotEqual(t, oldUser.Email, updatedUser.Email)
	require.Equal(t, newEmail, updatedUser.Email)
	require.NotEqual(t, oldUser.FullName, updatedUser.FullName)
	require.Equal(t, newFullName, updatedUser.FullName)
}
