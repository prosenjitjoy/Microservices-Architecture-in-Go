package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"main/discovery"
	"main/discovery/consul"
	"main/movie/controller"
	metadatagateway "main/movie/gateway/metadata/grpc"
	ratinggateway "main/movie/gateway/rating/grpc"
	grpchandler "main/movie/handler/grpc"
	"main/rpc"
	"main/utils"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const serviceName = "movie"

func main() {
	var config string
	flag.StringVar(&config, "config", ".env", "Configuration path")
	flag.Parse()
	cfg := utils.LoadConfig(config)

	log.Println("Starting the movie service on port", cfg.MoviePort)
	registry, err := consul.NewRegistry(cfg.ConsulURL)
	if err != nil {
		log.Fatal("failed to connect consul registry:", err)
	}

	ctx := context.Background()
	instanceID := discovery.GenerateInstanceID(serviceName)
	hostPort := fmt.Sprintf("%s:%d", cfg.Host, cfg.MoviePort)
	if err := registry.Register(ctx, instanceID, serviceName, hostPort); err != nil {
		log.Fatal("failed to register movie service:", err)
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

	metadataGateway := metadatagateway.New(registry)
	ratingGateway := ratinggateway.New(registry)
	svc := controller.New(ratingGateway, metadataGateway)
	h := grpchandler.New(svc)

	listener, err := net.Listen("tcp", hostPort)
	if err != nil {
		log.Fatalf("failed to listen on %s: %v", hostPort, err)
	}

	server := grpc.NewServer()
	reflection.Register(server)
	rpc.RegisterMovieServiceServer(server, h)
	if err := server.Serve(listener); err != nil {
		log.Fatal("Failed to start gRPC server:", err)
	}
}
