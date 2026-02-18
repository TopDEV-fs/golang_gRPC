package domain

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMoneyCalculations(t *testing.T) {
	price, err := NewMoney(1999, 100)
	require.NoError(t, err)

	discountFraction := big.NewRat(20, 100)
	discount, err := price.Mul(discountFraction)
	require.NoError(t, err)

	effective, err := price.Sub(discount)
	require.NoError(t, err)

	assert.Equal(t, "15.99", effective.String())
}
