package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"main/discovery"
	"main/discovery/consul"
	"main/metadata/controller"
	grpchandler "main/metadata/handler/grpc"
	"main/rpc"

	// "main/metadata/handler/api"
	"main/metadata/repository/memory"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const serviceName = "metadata"

func main() {
	var port int
	flag.IntVar(&port, "port", 8081, "API handler port")
	flag.Parse()

	log.Println("Starting the movie metadata service on port", port)
	registry, err := consul.NewRegistry("localhost:8500")
	if err != nil {
		log.Fatal("failed to connect consul registry:", err)
	}

	ctx := context.Background()
	instanceID := discovery.GenerateInstanceID(serviceName)
	hostPort := fmt.Sprintf("localhost:%d", port)
	if err := registry.Register(ctx, instanceID, serviceName, hostPort); err != nil {
		log.Fatal("failed to register metadata service:", err)
	}

	go func() {
		for {
			if err := registry.ReportHealthyState(instanceID, serviceName); err != nil {
				log.Println("Failed to report healthy state:", err)
			}
			time.Sleep(1 * time.Second)
		}
	}()
	defer registry.Deregister(ctx, instanceID, serviceName)

	repo := memory.New()
	svc := controller.New(repo)

	// h := api.New(svc)
	// http.Handle("/metadata", http.HandlerFunc(h.Handle))
	// if err := http.ListenAndServe(hostPort, nil); err != nil {
	// 	log.Fatal("Failed to start the server:", err)
	// }

	h := grpchandler.New(svc)
	listener, err := net.Listen("tcp", hostPort)
	if err != nil {
		log.Fatalf("failed to listen on %s: %v", hostPort, err)
	}

	server := grpc.NewServer()
	reflection.Register(server)
	rpc.RegisterMetadataServiceServer(server, h)
	if err := server.Serve(listener); err != nil {
		log.Fatal("Failed to start gRPC server:", err)
	}
}
