package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"main/database/db"
	"main/discovery"
	"main/discovery/consul"
	grpchandler "main/rating/handler/grpc"
	"main/rating/repository/postgres"
	"main/rating/service"
	"main/rpc"
	"main/utils"
	"net"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const serviceName = "rating"

func main() {
	var config string
	flag.StringVar(&config, "config", ".env", "Configuration path")
	flag.Parse()
	cfg := utils.LoadConfig(config)

	log.Println("Starting the rating service on port", cfg.RatingPort)
	registry, err := consul.NewRegistry(cfg.ConsulURL)
	if err != nil {
		log.Fatal("faild to connect to consul registry:", err)
	}

	ctx := context.Background()
	instanceID := discovery.GenerateInstanceID(serviceName)
	hostPort := fmt.Sprintf("%s:%d", cfg.Host, cfg.RatingPort)
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

	conn, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	store := db.NewStore(conn)
	repo := postgres.New(store)
	svc := service.New(repo, cfg)
	h := grpchandler.New(svc)

	go func() {
		if err := svc.StartConsume(ctx); err != nil {
			log.Fatal(err)
		}
	}()

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
