package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"main/database/db"
	"main/discovery"
	"main/discovery/consul"
	grpchandler "main/metadata/handler/grpc"
	"main/metadata/repository/postgres"
	"main/metadata/service"
	"main/rpc"
	"main/utils"
	"net"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const serviceName = "metadata"

func main() {
	var config string
	flag.StringVar(&config, "config", ".env", "Configuration path")
	flag.Parse()
	cfg := utils.LoadConfig(config)

	log.Println("Starting the movie metadata service on port", cfg.MetadataPort)
	registry, err := consul.NewRegistry(cfg.ConsulURL)
	if err != nil {
		log.Fatal("failed to connect consul registry:", err)
	}

	ctx := context.Background()
	instanceID := discovery.GenerateInstanceID(serviceName)
	hostPort := fmt.Sprintf("%s:%d", cfg.Host, cfg.MetadataPort)
	fmt.Println(hostPort)
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

	conn, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	store := db.NewStore(conn)
	repo := postgres.New(store)
	svc := service.New(repo)
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
