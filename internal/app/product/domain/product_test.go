package domain

import (
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplyDiscountToInactiveProduct(t *testing.T) {
	price, _ := NewMoney(1000, 100)
	product, err := NewProduct("p1", "A", "B", "C", price, time.Now().UTC())
	require.NoError(t, err)

	discount, err := NewDiscount(big.NewRat(10, 1), time.Now().Add(-time.Hour), time.Now().Add(time.Hour))
	require.NoError(t, err)

	err = product.ApplyDiscount(discount, time.Now())
	assert.ErrorIs(t, err, ErrProductNotActive)
}
