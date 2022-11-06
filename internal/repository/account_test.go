package repository_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/thirochas/simplebank-golang-api/internal/repository"
)

func createAccountByUser(t *testing.T, user repository.User, currency string) repository.Account {
	createAccountParams := repository.CreateAccountParams{
		Owner:    user.Username,
		Balance:  10,
		Currency: currency,
	}

	accountCreated, err := testQueries.CreateAccount(context.Background(), createAccountParams)
	require.NoError(t, err)
	require.NotEmpty(t, accountCreated)

	require.Equal(t, createAccountParams.Owner, accountCreated.Owner)
	require.Equal(t, createAccountParams.Balance, accountCreated.Balance)
	require.Equal(t, createAccountParams.Currency, accountCreated.Currency)

	require.NotZero(t, accountCreated.ID)
	require.NotZero(t, accountCreated.CreatedAt)

	return accountCreated
}

func createAccount(t *testing.T) repository.Account {
	user := createUser(t)

	createAccountParams := repository.CreateAccountParams{
		Owner:    user.Username,
		Balance:  10,
		Currency: "BRL",
	}

	accountCreated, err := testQueries.CreateAccount(context.Background(), createAccountParams)
	require.NoError(t, err)
	require.NotEmpty(t, accountCreated)

	require.Equal(t, createAccountParams.Owner, accountCreated.Owner)
	require.Equal(t, createAccountParams.Balance, accountCreated.Balance)
	require.Equal(t, createAccountParams.Currency, accountCreated.Currency)

	require.NotZero(t, accountCreated.ID)
	require.NotZero(t, accountCreated.CreatedAt)

	return accountCreated
}

func TestCreateAccount(t *testing.T) {
	createAccount(t)
}

func TestGetAccountByID(t *testing.T) {
	account := createAccount(t)

	accountFounded, err := testQueries.GetAccount(context.Background(), account.ID)
	require.NoError(t, err)
	require.NotEmpty(t, accountFounded)

	require.Equal(t, account.Owner, accountFounded.Owner)
	require.Equal(t, account.Balance, accountFounded.Balance)
	require.Equal(t, account.Currency, accountFounded.Currency)

	require.NotZero(t, accountFounded.ID)
	require.NotZero(t, accountFounded.CreatedAt)
}

func TestUpdateAccount(t *testing.T) {
	account := createAccount(t)

	updateAccountParams := repository.UpdateAccountParams{
		ID:      account.ID,
		Balance: 500000,
	}

	accountUpdated, err := testQueries.UpdateAccount(context.Background(), updateAccountParams)
	require.NoError(t, err)
	require.NotEmpty(t, accountUpdated)

	require.Equal(t, updateAccountParams.ID, accountUpdated.ID)
	require.Equal(t, updateAccountParams.Balance, accountUpdated.Balance)
}

func TestDelete(t *testing.T) {
	account := createAccount(t)

	err := testQueries.DeleteAccount(context.Background(), account.ID)
	require.NoError(t, err)
}

func TestListAccounts(t *testing.T) {
	user := createUser(t)
	for i := 0; i < 10; i++ {
		createAccountByUser(t, user, fmt.Sprintf("BRL-%d", i))
	}

	arg := repository.ListAccountsParams{
		Owner:  user.Username,
		Limit:  5,
		Offset: 5,
	}

	accounts, err := testQueries.ListAccounts(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, accounts, 5)

	for _, account := range accounts {
		require.NotEmpty(t, account)
	}
}
