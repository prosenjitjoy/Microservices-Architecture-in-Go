package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"main/discovery"
	"main/discovery/consul"
	metadatagateway "main/movie/gateway/metadata/grpc"
	ratinggateway "main/movie/gateway/rating/grpc"
	grpchandler "main/movie/handler/grpc"
	"main/movie/service"
	"main/rpc"
	"main/tracing"
	"main/util"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	// "github.com/grpc-ecosystem/go-grpc-middleware/ratelimit"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	serviceName = "movie"
	limit       = 100
	burst       = 100
)

func main() {
	var config string
	flag.StringVar(&config, "config", ".env", "Configuration path")
	flag.Parse()
	cfg := util.LoadConfig(config)

	if cfg.Environment == "dev" {
		var logger = slog.New(slog.NewTextHandler(os.Stdout, nil)).With("service_name", serviceName)
		slog.SetDefault(logger)
	} else {
		var logger = slog.New(slog.NewJSONHandler(os.Stdout, nil)).With("service_name", serviceName)
		slog.SetDefault(logger)
	}

	slog.Info("Starting the movie service on port", slog.Int("port", cfg.MoviePort))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tp, err := tracing.NewJaegerProvider(cfg.JaegerURL, serviceName)
	if err != nil {
		slog.Error("failed to initialize Jaeger provider:", slog.String("error", err.Error()))
		return
	}
	defer func() {
		if err := tp.Shutdown(ctx); err != nil {
			slog.Error("failed to shutdown tracing provider:", slog.String("error", err.Error()))
			return
		}
	}()

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	registry, err := consul.NewRegistry(cfg.ConsulURL)
	if err != nil {
		slog.Error("failed to connect consul registry:", slog.String("error", err.Error()))
		return
	}

	instanceID := discovery.GenerateInstanceID(serviceName)
	hostPort := fmt.Sprintf("%s:%d", cfg.Host, cfg.MoviePort)
	if err := registry.Register(ctx, instanceID, serviceName, hostPort); err != nil {
		slog.Error("failed to register movie service:", slog.String("error", err.Error()))
		return
	}

	go func() {
		for {
			if err := registry.ReportHealthyState(instanceID, serviceName); err != nil {
				slog.Info("Failed to report healthy state:", slog.String("error", err.Error()))
			}
			time.Sleep(1 * time.Second)
		}
	}()
	defer registry.Deregister(ctx, instanceID, serviceName)

	metadataGateway := metadatagateway.New(registry)
	ratingGateway := ratinggateway.New(registry)
	svc := service.New(ratingGateway, metadataGateway)
	h := grpchandler.New(svc)

	listener, err := net.Listen("tcp", hostPort)
	if err != nil {
		slog.Error("failed to listen on:", slog.String("host", hostPort), slog.String("error", err.Error()))
		return
	}

	// l := util.NewLimiter(limit, burst)
	server := grpc.NewServer(
		// grpc.UnaryInterceptor(ratelimit.UnaryServerInterceptor(l)),
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
	)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		s := <-sigChan
		cancel()
		slog.Info("Received signal", s)
		slog.Info("attempting graceful stutdown")
		server.GracefulStop()
		slog.Info("Gracefully stopped the gRPC server")
	}()

	reflection.Register(server)
	rpc.RegisterMovieServiceServer(server, h)
	if err := server.Serve(listener); err != nil {
		slog.Error("Failed to start gRPC server:", slog.String("error", err.Error()))
		return
	}
}
