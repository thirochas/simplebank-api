package infra

import (
	"database/sql"
	"github.com/thirochas/simplebank-golang-api/internal/util"
	"log"

	_ "github.com/lib/pq"
)

func CreateConnection() *sql.DB {
	config, err := util.LoadFromFiles("../..")
	if err != nil {
		log.Fatalf("cannot load env configs: %v", err)
	}

	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatalf("cannot connect to db: %v", err)
	}

	return conn
}
