package main

import (
	"github.com/thirochas/simplebank-golang-api/internal/api"
	"github.com/thirochas/simplebank-golang-api/internal/infra"
	"github.com/thirochas/simplebank-golang-api/internal/repository"
	"github.com/thirochas/simplebank-golang-api/internal/util"
	"log"
)

func main() {
	config, err := util.LoadFromFiles(".")
	if err != nil {
		log.Fatalf("cannot load env configs: %v", err)
	}

	conn := infra.CreateConnection()

	store := repository.NewStore(conn)
	server := api.NewServer(store, config)

	server.Start(config.ServerAddress)
}
