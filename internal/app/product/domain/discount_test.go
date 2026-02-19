package domain

import (
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDiscountValidation(t *testing.T) {
	now := time.Now().UTC()

	_, err := NewDiscount(big.NewRat(0, 1), now, now.Add(time.Hour))
	assert.ErrorIs(t, err, ErrInvalidDiscountPercent)

	_, err = NewDiscount(big.NewRat(101, 1), now, now.Add(time.Hour))
	assert.ErrorIs(t, err, ErrInvalidDiscountPercent)

	_, err = NewDiscount(big.NewRat(10, 1), now, now)
	assert.ErrorIs(t, err, ErrInvalidDiscountPeriod)
}
