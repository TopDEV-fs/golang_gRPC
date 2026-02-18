package list_products

import (
	"context"
	"math/big"
	"time"

	"github.com/example/product-catalog-service/internal/app/product/contracts"
	"github.com/example/product-catalog-service/internal/app/product/domain"
)

type Query struct {
	readModel contracts.ProductReadModel
}

func New(readModel contracts.ProductReadModel) *Query {
	return &Query{readModel: readModel}
}

func (q *Query) Execute(ctx context.Context, category string, pageSize int32, pageToken string) (*Result, error) {
	rows, err := q.readModel.ListActive(ctx, category, pageSize, pageToken)
	if err != nil {
		return nil, err
	}
	items := make([]ProductDTO, 0, len(rows.Items))
	now := time.Now().UTC()
	for _, row := range rows.Items {
		basePrice, err := domain.NewMoney(row.BasePriceNumerator, row.BasePriceDenominator)
		if err != nil {
			continue
		}
		effective := basePrice
		if row.DiscountPercent != "" {
			if pct, ok := new(big.Rat).SetString(row.DiscountPercent); ok {
				d, err := domain.NewDiscount(pct, time.Unix(row.DiscountStartUnix, 0).UTC(), time.Unix(row.DiscountEndUnix, 0).UTC())
				if err == nil && d.IsValidAt(now) {
					dv, _ := basePrice.Mul(d.Fraction())
					effective, _ = basePrice.Sub(dv)
				}
			}
		}
		items = append(items, ProductDTO{
			ID:             row.ID,
			Name:           row.Name,
			Description:    row.Description,
			Category:       row.Category,
			Status:         row.Status,
			BasePrice:      basePrice.String(),
			EffectivePrice: effective.String(),
		})
	}

	return &Result{Items: items, NextPageToken: rows.NextPageToken}, nil
}
