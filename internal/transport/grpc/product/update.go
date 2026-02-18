package product

import (
	"context"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"github.com/example/product-catalog-service/internal/app/product/usecases/activate_product"
	"github.com/example/product-catalog-service/internal/app/product/usecases/apply_discount"
	"github.com/example/product-catalog-service/internal/app/product/usecases/update_product"
	productv1 "github.com/example/product-catalog-service/proto/product/v1"
)

func (h *Handler) UpdateProduct(ctx context.Context, req *productv1.UpdateProductRequest) (*productv1.UpdateProductReply, error) {
	if req.ProductId == "" {
		return nil, status.Error(codes.InvalidArgument, "product_id is required")
	}
	err := h.container.Commands.UpdateProduct.Execute(ctx, update_product.Request{
		ProductID:    req.ProductId,
		Name:         req.Name,
		Description:  req.Description,
		Category:     req.Category,
	})
	if err != nil {
		return nil, mapDomainError(err)
	}
	return &productv1.UpdateProductReply{}, nil
}

func (h *Handler) ActivateProduct(ctx context.Context, req *productv1.ActivateProductRequest) (*productv1.ActivateProductReply, error) {
	err := h.container.Commands.Activate.Activate(ctx, activate_product.ActivateRequest{ProductID: req.ProductId})
	if err != nil {
		return nil, mapDomainError(err)
	}
	return &productv1.ActivateProductReply{}, nil
}

func (h *Handler) DeactivateProduct(ctx context.Context, req *productv1.DeactivateProductRequest) (*productv1.DeactivateProductReply, error) {
	err := h.container.Commands.Activate.Deactivate(ctx, activate_product.DeactivateRequest{ProductID: req.ProductId})
	if err != nil {
		return nil, mapDomainError(err)
	}
	return &productv1.DeactivateProductReply{}, nil
}

func (h *Handler) ApplyDiscount(ctx context.Context, req *productv1.ApplyDiscountRequest) (*productv1.ApplyDiscountReply, error) {
	err := h.container.Commands.Discount.Apply(ctx, apply_discount.ApplyRequest{
		ProductID:    req.ProductId,
		Percent:      req.Percent,
		StartDateUTC: time.Unix(req.StartDateUnix, 0).UTC(),
		EndDateUTC:   time.Unix(req.EndDateUnix, 0).UTC(),
	})
	if err != nil {
		return nil, mapDomainError(err)
	}
	return &productv1.ApplyDiscountReply{}, nil
}

func (h *Handler) RemoveDiscount(ctx context.Context, req *productv1.RemoveDiscountRequest) (*productv1.RemoveDiscountReply, error) {
	err := h.container.Commands.Discount.Remove(ctx, apply_discount.RemoveRequest{ProductID: req.ProductId})
	if err != nil {
		return nil, mapDomainError(err)
	}
	return &productv1.RemoveDiscountReply{}, nil
}
