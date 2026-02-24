package domain

import (
	"math/big"
)

// Money is an immutable value object representing a non-negative monetary amount
// stored as an exact rational number (*big.Rat) to avoid floating-point precision
// errors during arithmetic operations.
type Money struct {
	amount *big.Rat
}

// NewMoney creates a Money value from an integer numerator/denominator pair.
// Returns ErrInvalidPrice if the denominator is zero or the resulting value is negative.
func NewMoney(num, den int64) (*Money, error) {
	if den == 0 {
		return nil, ErrInvalidPrice
	}
	r := big.NewRat(num, den)
	if r.Sign() < 0 {
		return nil, ErrInvalidPrice
	}
	return &Money{amount: r}, nil
}

// NewMoneyFromRat creates a Money value from an existing *big.Rat. The value is
// copied so the caller may safely reuse or mutate the original rational.
func NewMoneyFromRat(r *big.Rat) (*Money, error) {
	if r == nil || r.Sign() < 0 {
		return nil, ErrInvalidPrice
	}
	return &Money{amount: new(big.Rat).Set(r)}, nil
}

// Amount returns a copy of the underlying rational value.
func (m *Money) Amount() *big.Rat {
	if m == nil || m.amount == nil {
		return big.NewRat(0, 1)
	}
	return new(big.Rat).Set(m.amount)
}

// Numerator returns the numerator component of the underlying rational in lowest terms.
func (m *Money) Numerator() int64 {
	return m.Amount().Num().Int64()
}

// Denominator returns the denominator component of the underlying rational in lowest terms.
func (m *Money) Denominator() int64 {
	return m.Amount().Denom().Int64()
}

// Sub subtracts other from m and returns the result.
// Returns ErrInvalidPrice if either operand is nil or the result is negative.
func (m *Money) Sub(other *Money) (*Money, error) {
	if m == nil || other == nil {
		return nil, ErrInvalidPrice
	}
	res := new(big.Rat).Sub(m.Amount(), other.Amount())
	if res.Sign() < 0 {
		return nil, ErrInvalidPrice
	}
	return NewMoneyFromRat(res)
}

// Mul multiplies m by the rational r and returns the result.
// Returns ErrInvalidPrice if either operand is nil or the result is negative.
func (m *Money) Mul(r *big.Rat) (*Money, error) {
	if m == nil || r == nil {
		return nil, ErrInvalidPrice
	}
	res := new(big.Rat).Mul(m.Amount(), r)
	if res.Sign() < 0 {
		return nil, ErrInvalidPrice
	}
	return NewMoneyFromRat(res)
}

// String returns the amount formatted as a 2-decimal-place string, e.g. "19.99".
func (m *Money) String() string {
	return m.Amount().FloatString(2)
}
