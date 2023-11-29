create-consul:
	podman run --name dev-consul -p 8500:8500 -p 8600:8600/udp -d hashicorp/consul:latest agent -server -ui -node=server-1 -bootstrap-expect=1 -client=0.0.0.0

delete-consul:
	podman rm -f dev-consul
	podman volume prune

.PHONY: create-consul delete-consul