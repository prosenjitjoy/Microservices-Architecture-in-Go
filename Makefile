create-consul:
	podman run --name dev-consul -p 8500:8500 -p 8600:8600/udp -d hashicorp/consul:latest agent -server -ui -node=server-1 -bootstrap-expect=1 -client=0.0.0.0

delete-consul:
	podman rm -f dev-consul
	podman volume prune

proto-generate:
	rm -rf rpc/*
	protoc --proto_path=proto --go_out=rpc --go_opt=paths=source_relative --go-grpc_out=rpc --go-grpc_opt=paths=source_relative proto/*.proto

create-pulsar:
	podman run --name dev-pulsar -p 6650:6650 -p 8080:8080 -d apachepulsar/pulsar:3.1.1 bin/pulsar standalone

delete-pulsar:
	podman rm -f dev-pulsar
	podman volume prune

.PHONY: create-consul delete-consul proto-generate create-pulsar delete-pulsar