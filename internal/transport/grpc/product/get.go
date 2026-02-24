package product

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	productv1 "github.com/example/product-catalog-service/proto/product/v1"
)

// GetProduct validates the request and delegates to the GetProduct query.
func (h *Handler) GetProduct(ctx context.Context, req *productv1.GetProductRequest) (*productv1.GetProductReply, error) {
	if req.ProductId == "" {
		return nil, status.Error(codes.InvalidArgument, "product_id is required")
	}
	dto, err := h.container.Queries.GetProduct.Execute(ctx, req.ProductId)
	if err != nil {
		return nil, mapDomainError(err)
	}
	resp := &productv1.GetProductReply{
		ProductId:      dto.ID,
		Name:           dto.Name,
		Description:    dto.Description,
		Category:       dto.Category,
		Status:         dto.Status,
		BasePrice:      dto.BasePrice,
		EffectivePrice: dto.EffectivePrice,
	}
	if dto.Discount != nil {
		resp.Discount = &productv1.Discount{
			Percent:       dto.Discount.Percent,
			StartDateUnix: dto.Discount.StartDateUnix,
			EndDateUnix:   dto.Discount.EndDateUnix,
		}
	}
	return resp, nil
}
