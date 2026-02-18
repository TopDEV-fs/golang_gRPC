package product

import (
	"context"

	productv1 "github.com/example/product-catalog-service/proto/product/v1"
)

func (h *Handler) ListProducts(ctx context.Context, req *productv1.ListProductsRequest) (*productv1.ListProductsReply, error) {
	result, err := h.container.Queries.ListProducts.Execute(ctx, req.Category, req.PageSize, req.PageToken)
	if err != nil {
		return nil, mapDomainError(err)
	}
	products := make([]*productv1.Product, 0, len(result.Items))
	for _, item := range result.Items {
		products = append(products, &productv1.Product{
			ProductId:      item.ID,
			Name:           item.Name,
			Description:    item.Description,
			Category:       item.Category,
			Status:         item.Status,
			BasePrice:      item.BasePrice,
			EffectivePrice: item.EffectivePrice,
		})
	}
	return &productv1.ListProductsReply{Products: products, NextPageToken: result.NextPageToken}, nil
}
