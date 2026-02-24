package product

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/example/product-catalog-service/internal/app/product/usecases/create_product"
	productv1 "github.com/example/product-catalog-service/proto/product/v1"
)

// CreateProduct validates the request and delegates to the CreateProduct use case.
func (h *Handler) CreateProduct(ctx context.Context, req *productv1.CreateProductRequest) (*productv1.CreateProductReply, error) {
	if req.Name == "" || req.Category == "" || req.BasePriceDenominator == 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid create request")
	}
	productID, err := h.container.Commands.CreateProduct.Execute(ctx, create_product.Request{
		Name:                 req.Name,
		Description:          req.Description,
		Category:             req.Category,
		BasePriceNumerator:   req.BasePriceNumerator,
		BasePriceDenominator: req.BasePriceDenominator,
	})
	if err != nil {
		return nil, mapDomainError(err)
	}
	return &productv1.CreateProductReply{ProductId: productID}, nil
}
