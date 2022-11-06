postgres:
	docker run --name postgres12 -p 5432:5432 -e POSTGRES_PASSWORD=123456 -e POSTGRES_USER=root -d postgres:12-alpine

createdb:
	docker exec -it postgres12 createdb --username=root --owner=root simples_bank_db

dropdb:
	docker exec -it postgres12 dropdb --username=root --owner=root simples_bank_db

migrateup:
	migrate -path db/migration -database "postgresql://root:123456@127.0.0.1:5432/simples_bank_db?sslmode=disable" -verbose up

migrateup1:
	migrate -path db/migration -database "postgresql://root:123456@127.0.0.1:5432/simples_bank_db?sslmode=disable" -verbose up 1

migratedown:
	migrate -path db/migration -database "postgresql://root:123456@127.0.0.1:5432/simples_bank_db?sslmode=disable" -verbose down

migratedown1:
	migrate -path db/migration -database "postgresql://root:123456@127.0.0.1:5432/simples_bank_db?sslmode=disable" -verbose down 1

sqlc:
	sqlc generate

build:
	go build ./...

test:
	go test -v -cover ./...

mock:
	mockgen -package mockdb -destination internal/repository/mock/store.go github.com/thirochas/simplebank-golang-api/internal/repository Store

.PHONY: postgres createdb dropdb migrateup migrateup1 migratedown migratedown1