package fake

import (
	"github.com/thirochas/simplebank-golang-api/internal/repository"
	"github.com/thirochas/simplebank-golang-api/internal/util"
)

func Account() repository.Account {
	return repository.Account{
		ID:       50,
		Owner:    "user",
		Balance:  70000,
		Currency: "BRL",
	}
}

func Accounts() []repository.Account {
	return []repository.Account{
		{
			ID:       50,
			Owner:    "user",
			Balance:  70000,
			Currency: "BRL",
		},
		{
			ID:       51,
			Owner:    "user-2",
			Balance:  1000000,
			Currency: "BRL",
		},
	}
}

func User() (repository.User, error) {
	password, err := util.GenerateHashPassword("secret")
	if err != nil {
		return repository.User{}, err
	}
	user := repository.User{
		Username:       util.RandomString(10),
		HashedPassword: password,
		FullName:       util.RandomString(20),
		Email:          util.RandomEmail(),
	}

	return user, nil
}
