package list_products

import (
	"context"
	"math/big"
	"time"

	"github.com/example/product-catalog-service/internal/app/product/contracts"
	"github.com/example/product-catalog-service/internal/app/product/domain"
)

// Query executes the ListProducts read query.
type Query struct {
	readModel contracts.ProductReadModel
}

// New returns a new ListProducts Query backed by the given read model.
func New(readModel contracts.ProductReadModel) *Query {
	return &Query{readModel: readModel}
}

// Execute fetches a paginated list of active products, optionally filtered by
// category, and resolves the effective price for each item. Returns a Result
// containing the page items and a token for the next page.
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
