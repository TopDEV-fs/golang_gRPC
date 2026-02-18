package services

import (
	"time"

	"github.com/example/product-catalog-service/internal/app/product/domain"
)

type PricingCalculator struct{}

func NewPricingCalculator() *PricingCalculator {
	return &PricingCalculator{}
}

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
