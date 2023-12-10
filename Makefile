include .env

create-consul:
	podman run --name consul --hostname consul -p 8500:8500 -p 8600:8600/udp -d hashicorp/consul:latest agent -server -ui -node=server-1 -bootstrap-expect=1 -client=0.0.0.0

delete-consul:
	podman rm -f consul
	podman volume prune

proto-generate:
	rm -rf rpc/*
	protoc --proto_path=proto --go_out=rpc --go_opt=paths=source_relative --go-grpc_out=rpc --go-grpc_opt=paths=source_relative proto/*.proto

create-pulsar:
	podman run --name pulsar --hostname pulsar -p 6650:6650 -p 8080:8080 -d apachepulsar/pulsar:3.1.1 bin/pulsar standalone

delete-pulsar:
	podman rm -f pulsar
	podman volume prune

create-postgres:
	podman run --name postgres --hostname postgres -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=moviedb -p 5432:5432 -d postgres:latest

delete-postgres:
	podman rm -f postgres
	podman volume prune

create-migration:
	migrate create -ext sql -dir database/migration -seq $(name)

migrate-up:
	migrate -database ${DATABASE_URL} -path database/migration -verbose up

migrate-down:
	migrate -database ${DATABASE_URL} -path database/migration -verbose down

create-jaeger:
	podman run --name jaeger --hostname jaeger -e COLLECTOR_OTLP_ENABLED=true -p 6831:6831/udp -p 6832:6832/udp -p 5778:5778 -p 16686:16686 -p 4317:4317 -p 4318:4318 -p 14250:14250 -p 14268:14268 -p 14269:14269 -p 9411:9411 -d jaegertracing/all-in-one:1.52

delete-jaeger:
	podman rm -f jaeger
	podman volume prune

create-prometheus:
	podman run --name prometheus --hostname prometheus --network slirp4netns:allow_host_loopback=true -p 9090:9090 -v ./configs:/etc/prometheus -d prom/prometheus:latest

delete-prometheus:
	podman rm -f prometheus
	podman volume prune

create-alertmanager:
	podman run --name alertmanager --hostname alertmanager -p 9093:9093 -v ./configs:/etc/alertmanager -d prom/alertmanager:latest --config.file=/etc/alertmanager/alertmanager.yml

delete-alertmanager:
	podman rm -f alertmanager
	podman volume prune

create-grafana:
	podman run --name grafana --hostname grafana -p 3000:3000 -d grafana/grafana-oss:latest

delete-grafana:
	podman rm -f grafana
	podman volume prune

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

cpu-profiling:
	go tool pprof cpu.pprof

mem-profiling:
	go tool pprof mem.pprof

.PHONY: create-consul delete-consul proto-generate create-pulsar delete-pulsar create-postgres delete-postgres create-migration migrate-up migrate-down create-jaeger delete-jaeger create-prometheus delete-prometheus create-alertmanager delete-alertmanager create-grafana delete-grafana generate-dbdocs generate-schema generate-sqlc generate-image generate-mock run-test cpu-profiling mem-profiling