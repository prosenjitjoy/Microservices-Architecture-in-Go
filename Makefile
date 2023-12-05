include .env

create-consul:
	podman run --name consul -p 8500:8500 -p 8600:8600/udp -d hashicorp/consul:latest agent -server -ui -node=server-1 -bootstrap-expect=1 -client=0.0.0.0

delete-consul:
	podman rm -f consul
	podman volume prune

proto-generate:
	rm -rf rpc/*
	protoc --proto_path=proto --go_out=rpc --go_opt=paths=source_relative --go-grpc_out=rpc --go-grpc_opt=paths=source_relative proto/*.proto

create-pulsar:
	podman run --name pulsar -p 6650:6650 -p 8080:8080 -d apachepulsar/pulsar:3.1.1 bin/pulsar standalone

delete-pulsar:
	podman rm -f pulsar
	podman volume prune

create-postgres:
	podman run --name postgres -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=moviedb -p 5432:5432 -d postgres:latest

delete-postgres:
	podman rm -f postgres
	podman volume prune

create-migration:
	migrate create -ext sql -dir database/migration -seq $(name)

migrate-up:
	migrate -database ${DATABASE_URL} -path database/migration -verbose up

migrate-down:
	migrate -database ${DATABASE_URL} -path database/migration -verbose down

generate-dbdocs:
	dbdocs build database/doc/db.dbml

generate-schema:
	dbml2sql --postgres -o database/doc/schema.sql database/doc/db.dbml

generate-sqlc:
	sqlc generate

generate-image:
	podman build --tag=metadata --target=metadata .
	podman build --tag=rating --target=rating .
	podman build --tag=movie --target=movie .

generate-mock:
	mockgen -package mockdb -destination database/mockdb/store.go main/database/db Store

run-test:
	go test ./...

.PHONY: create-consul delete-consul proto-generate create-pulsar delete-pulsar create-postgres delete-postgres create-migration migrate-up migrate-down generate-dbdocs generate-schema generate-sqlc generate-image generate-mock run-test