package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"main/database/db"
	"main/discovery"
	"main/discovery/consul"
	"main/metadata/controller"
	grpchandler "main/metadata/handler/grpc"
	"main/metadata/repository/postgres"
	"main/rpc"
	"main/utils"

	// "main/metadata/handler/api"
	// "main/metadata/repository/memory"
	"net"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const serviceName = "metadata"

func main() {
	var port int
	var config string
	flag.IntVar(&port, "port", 8081, "API handler port")
	flag.StringVar(&config, "config", ".env", "Configuration path")
	flag.Parse()

	cfg := utils.LoadConfig(config)
	_ = cfg

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

	// repo := memory.New()

	conn, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}
	store := db.NewStore(conn)

	repo := postgres.New(store)
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
