package product

import (
	"github.com/example/product-catalog-service/internal/services"
	productv1 "github.com/example/product-catalog-service/proto/product/v1"
)

type Handler struct {
	productv1.UnimplementedProductServiceServer
	container *services.Container
}

func NewHandler(container *services.Container) *Handler {
	return &Handler{container: container}
}
