package services

import (
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/example/product-catalog-service/internal/app/product/domain"
)

func TestEffectivePriceWithActiveDiscount(t *testing.T) {
	price, err := domain.NewMoney(1000, 100)
	require.NoError(t, err)

	now := time.Now().UTC()
	product, err := domain.NewProduct("p1", "Book", "Desc", "books", price, now)
	require.NoError(t, err)
	require.NoError(t, product.Activate(now))

	d, err := domain.NewDiscount(big.NewRat(20, 1), now.Add(-time.Minute), now.Add(time.Minute))
	require.NoError(t, err)
	require.NoError(t, product.ApplyDiscount(d, now))

	calc := NewPricingCalculator()
	effective, err := calc.EffectivePrice(product, now)
	require.NoError(t, err)
	assert.Equal(t, "8.00", effective.String())
}
