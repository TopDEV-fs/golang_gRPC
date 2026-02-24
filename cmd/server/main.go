// Command server starts the Product Catalog gRPC server.
//
// Required environment variables:
//   - SPANNER_DATABASE – fully-qualified Spanner database path,
//     e.g. projects/my-project/instances/my-instance/databases/my-db
//
// Optional environment variables:
//   - GRPC_ADDR – TCP address to listen on (default :50051)
package main

import (
	"context"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cloud.google.com/go/spanner"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/example/product-catalog-service/internal/services"
	grpcproduct "github.com/example/product-catalog-service/internal/transport/grpc/product"
	productv1 "github.com/example/product-catalog-service/proto/product/v1"
)

const gracefulStopTimeout = 15 * time.Second

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(log)

	dsn := os.Getenv("SPANNER_DATABASE")
	if dsn == "" {
		slog.Error("SPANNER_DATABASE env is required")
		os.Exit(1)
	}
	addr := os.Getenv("GRPC_ADDR")
	if addr == "" {
		addr = ":50051"
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	spannerClient, err := spanner.NewClient(ctx, dsn)
	if err != nil {
		slog.Error("failed to create Spanner client", "error", err)
		os.Exit(1)
	}
	defer spannerClient.Close()

	container := services.BuildContainer(spannerClient)
	handler := grpcproduct.NewHandler(container)

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		slog.Error("failed to listen", "addr", addr, "error", err)
		os.Exit(1)
	}

	srv := grpc.NewServer()
	productv1.RegisterProductServiceServer(srv, handler)
	// Enable server reflection so tooling like grpcurl can introspect the API.
	reflection.Register(srv)

	slog.Info("gRPC server starting", "addr", addr)

	serveErr := make(chan error, 1)
	go func() {
		serveErr <- srv.Serve(lis)
	}()

	select {
	case <-ctx.Done():
		slog.Info("shutdown signal received, draining connections", "timeout", gracefulStopTimeout)
		stopped := make(chan struct{})
		go func() {
			srv.GracefulStop()
			close(stopped)
		}()
		select {
		case <-stopped:
			slog.Info("server stopped gracefully")
		case <-time.After(gracefulStopTimeout):
			slog.Warn("graceful stop timed out, forcing shutdown")
			srv.Stop()
		}
	case err := <-serveErr:
		if err != nil {
			slog.Error("server exited unexpectedly", "error", err)
			os.Exit(1)
		}
	}
}
