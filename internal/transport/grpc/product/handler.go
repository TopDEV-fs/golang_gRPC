// Package product implements the gRPC transport layer for the ProductService.
// Each RPC method is responsible only for request validation, delegation to the
// application layer, and mapping results back to proto messages. Business logic
// must never be placed in this package.
package product

import (
	"github.com/example/product-catalog-service/internal/services"
	productv1 "github.com/example/product-catalog-service/proto/product/v1"
)

// Handler implements productv1.ProductServiceServer and delegates every RPC
// to the appropriate command or query in the application Container.
type Handler struct {
	productv1.UnimplementedProductServiceServer
	container *services.Container
}

// NewHandler returns a new Handler wired to the given application Container.
func NewHandler(container *services.Container) *Handler {
	return &Handler{container: container}
}
