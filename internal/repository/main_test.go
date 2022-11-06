package repository_test

import (
	"os"
	"testing"

	_ "github.com/lib/pq"
	"github.com/thirochas/simplebank-golang-api/internal/infra"
	"github.com/thirochas/simplebank-golang-api/internal/repository"
)

var testQueries *repository.Queries

func TestMain(m *testing.M) {
	testDB := infra.CreateConnection()

	testQueries = repository.New(testDB)

	os.Exit(m.Run())
}
