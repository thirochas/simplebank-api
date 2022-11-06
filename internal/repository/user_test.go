package repository_test

import (
	"context"
	"github.com/stretchr/testify/require"
	"github.com/thirochas/simplebank-golang-api/internal/repository"
	"github.com/thirochas/simplebank-golang-api/internal/util"
	"testing"
	"time"
)

func createUser(t *testing.T) repository.User {
	password, err := util.GenerateHashPassword("secret")
	require.NoError(t, err)

	userAccountParams := repository.CreateUserParams{
		Username:       util.RandomString(6),
		HashedPassword: password,
		FullName:       util.RandomString(10),
		Email:          util.RandomEmail(),
	}

	userCreated, err := testQueries.CreateUser(context.Background(), userAccountParams)
	require.NoError(t, err)
	require.NotEmpty(t, userCreated)

	require.Equal(t, userAccountParams.Username, userCreated.Username)
	require.Equal(t, userAccountParams.HashedPassword, userCreated.HashedPassword)
	require.Equal(t, userAccountParams.FullName, userCreated.FullName)
	require.Equal(t, userAccountParams.Email, userCreated.Email)

	return userCreated
}

func TestCreateUser(t *testing.T) {
	createUser(t)
}

func TestGetUserByID(t *testing.T) {
	user := createUser(t)

	accountFounded, err := testQueries.GetUser(context.Background(), user.Username)
	require.NoError(t, err)
	require.NotEmpty(t, accountFounded)

	require.Equal(t, user.Username, accountFounded.Username)
	require.Equal(t, user.HashedPassword, accountFounded.HashedPassword)
	require.Equal(t, user.FullName, accountFounded.FullName)
	require.Equal(t, user.Email, accountFounded.Email)

	require.WithinDuration(t, user.PasswordChangedAt, accountFounded.PasswordChangedAt, time.Second)
	require.WithinDuration(t, user.CreatedAt, accountFounded.CreatedAt, time.Second)
}
