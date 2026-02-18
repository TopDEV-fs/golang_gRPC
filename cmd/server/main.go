package main

import (
	"context"
	"log"
	"net"
	"os"

	"cloud.google.com/go/spanner"
	"google.golang.org/grpc"
	"github.com/example/product-catalog-service/internal/services"
	grpcproduct "github.com/example/product-catalog-service/internal/transport/grpc/product"
	productv1 "github.com/example/product-catalog-service/proto/product/v1"
)

func main() {
	dsn := os.Getenv("SPANNER_DATABASE")
	if dsn == "" {
		log.Fatal("SPANNER_DATABASE env is required")
	}
	addr := os.Getenv("GRPC_ADDR")
	if addr == "" {
		addr = ":50051"
	}

	ctx := context.Background()
	client, err := spanner.NewClient(ctx, dsn)
	if err != nil {
		log.Fatalf("spanner client: %v", err)
	}
	defer client.Close()

	container := services.BuildContainer(client)
	handler := grpcproduct.NewHandler(container)

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}
	server := grpc.NewServer()
	productv1.RegisterProductServiceServer(server, handler)

	log.Printf("gRPC server listening on %s", addr)
	if err := server.Serve(lis); err != nil {
		log.Fatalf("serve: %v", err)
	}
}
