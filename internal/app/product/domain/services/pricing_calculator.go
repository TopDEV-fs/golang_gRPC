// Package services contains stateless domain services that operate across
// multiple value objects or perform calculations that do not belong on a
// single aggregate or value object.
package services

import (
	"time"

	"github.com/example/product-catalog-service/internal/app/product/domain"
)

// PricingCalculator computes effective prices by applying active discounts.
type PricingCalculator struct{}

// NewPricingCalculator returns a new PricingCalculator.
func NewPricingCalculator() *PricingCalculator {
	return &PricingCalculator{}
}

// EffectivePrice returns the price of the product at the given instant after
// applying any currently active discount. If the product has no discount, or
// the discount is not valid at now, the base price is returned unchanged.
func (c *PricingCalculator) EffectivePrice(product *domain.Product, now time.Time) (*domain.Money, error) {
	if product.Discount() == nil {
		return product.BasePrice(), nil
	}
	if !product.Discount().IsValidAt(now) {
		return product.BasePrice(), nil
	}
	discountAmount, err := product.BasePrice().Mul(product.Discount().Fraction())
	if err != nil {
		return nil, err
	}
	return product.BasePrice().Sub(discountAmount)
}
