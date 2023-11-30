package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"main/discovery"
	"main/discovery/consul"
	"main/rating/controller"
	grpchandler "main/rating/handler/grpc"
	"main/rpc"
	"net"

	// "main/rating/handler/api"
	"main/rating/repository/memory"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const serviceName = "rating"

func main() {
	var port int
	flag.IntVar(&port, "port", 8082, "API handler port")
	flag.Parse()

	log.Println("Starting the rating service on port", port)
	registry, err := consul.NewRegistry("localhost:8500")
	if err != nil {
		log.Fatal("faild to connect to consul registry:", err)
	}

	ctx := context.Background()
	instanceID := discovery.GenerateInstanceID(serviceName)
	hostPort := fmt.Sprintf("localhost:%d", port)
	if err := registry.Register(ctx, instanceID, serviceName, hostPort); err != nil {
		log.Fatal("failed to register rating service:", err)
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
	// http.Handle("/rating", http.HandlerFunc(h.Handle))
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
	rpc.RegisterRatingServiceServer(server, h)
	if err := server.Serve(listener); err != nil {
		log.Fatal("Failed to start gRPC server:", err)
	}
}
