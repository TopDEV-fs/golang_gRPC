package get_product

import (
	"context"
	"math/big"
	"time"

	"github.com/example/product-catalog-service/internal/app/product/contracts"
	"github.com/example/product-catalog-service/internal/app/product/domain"
)

// Query executes the GetProduct read query.
type Query struct {
	readModel contracts.ProductReadModel
}

// New returns a new GetProduct Query backed by the given read model.
func New(readModel contracts.ProductReadModel) *Query {
	return &Query{readModel: readModel}
}

// Execute fetches the product with the given productID and computes the effective
// price by applying any currently active discount. Returns a fully populated
// ProductDTO or an error (including iterator.Done when not found).
func (q *Query) Execute(ctx context.Context, productID string) (*ProductDTO, error) {
	row, err := q.readModel.GetByID(ctx, productID)
	if err != nil {
		return nil, err
	}
	basePrice, err := domain.NewMoney(row.BasePriceNumerator, row.BasePriceDenominator)
	if err != nil {
		return nil, err
	}
	effective := basePrice
	var discountDTO *DiscountDTO
	if row.DiscountPercent != "" {
		discountPct, ok := new(big.Rat).SetString(row.DiscountPercent)
		if ok {
			d, err := domain.NewDiscount(discountPct, time.Unix(row.DiscountStartUnix, 0).UTC(), time.Unix(row.DiscountEndUnix, 0).UTC())
			if err == nil && d.IsValidAt(time.Now().UTC()) {
				discountValue, _ := basePrice.Mul(d.Fraction())
				effective, _ = basePrice.Sub(discountValue)
				discountDTO = &DiscountDTO{
					Percent:       row.DiscountPercent,
					StartDateUnix: row.DiscountStartUnix,
					EndDateUnix:   row.DiscountEndUnix,
				}
			}
		}
	}
	return &ProductDTO{
		ID:             row.ID,
		Name:           row.Name,
		Description:    row.Description,
		Category:       row.Category,
		Status:         row.Status,
		BasePrice:      basePrice.String(),
		EffectivePrice: effective.String(),
		Discount:       discountDTO,
		CreatedAtUnix:  row.CreatedAtUnix,
		UpdatedAtUnix:  row.UpdatedAtUnix,
	}, nil
}
