package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"main/database/db"
	"main/discovery"
	"main/discovery/consul"
	grpchandler "main/metadata/handler/grpc"
	"main/metadata/repository/postgres"
	"main/metadata/service"
	"main/rpc"
	"main/tracing"
	"main/util"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const serviceName = "metadata"

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

	// prometheus
	reg := prometheus.NewRegistry()
	counter := promauto.NewCounter(prometheus.CounterOpts{
		Namespace: serviceName,
		Name:      "service_started",
	})

	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{Registry: reg}))
	go func() {
		if err := http.ListenAndServe(fmt.Sprintf(":%d", cfg.MetadataMetricsPort), nil); err != nil {
			slog.Error("failed to start metrics handler:", slog.String("error", err.Error()))
			return
		}
	}()

	reg.MustRegister(counter)
	defer reg.Unregister(counter)
	counter.Inc()

	slog.Info("Starting the movie metadata service on port", slog.Int("port", cfg.MetadataPort))

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
	hostPort := fmt.Sprintf("%s:%d", cfg.Host, cfg.MetadataPort)
	fmt.Println(hostPort)
	if err := registry.Register(ctx, instanceID, serviceName, hostPort); err != nil {
		slog.Error("failed to register metadata service:", slog.String("error", err.Error()))
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

	conn, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		slog.Error("cannot connect to db:", slog.String("error", err.Error()))
		return
	}

	store := db.NewStore(conn)
	repo := postgres.New(store)
	svc := service.New(repo)
	h := grpchandler.New(svc)

	listener, err := net.Listen("tcp", hostPort)
	if err != nil {
		slog.Error("failed to listen on", slog.String("host", hostPort), slog.String("error", err.Error()))
		return
	}

	server := grpc.NewServer(grpc.StatsHandler(otelgrpc.NewServerHandler()))

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		s := <-sigChan
		cancel()
		slog.Info("Received signal", s)
		slog.Info("attempting graceful shutdown")
		server.GracefulStop()
		slog.Info("Gracefully stopped the gRPC server")
	}()

	reflection.Register(server)
	rpc.RegisterMetadataServiceServer(server, h)
	if err := server.Serve(listener); err != nil {
		slog.Error("Failed to start gRPC server:", slog.String("error", err.Error()))
		return
	}
}
