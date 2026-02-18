package domain

import (
	"math/big"
)

type Money struct {
	amount *big.Rat
}

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

func NewMoneyFromRat(r *big.Rat) (*Money, error) {
	if r == nil || r.Sign() < 0 {
		return nil, ErrInvalidPrice
	}
	return &Money{amount: new(big.Rat).Set(r)}, nil
}

func (m *Money) Amount() *big.Rat {
	if m == nil || m.amount == nil {
		return big.NewRat(0, 1)
	}
	return new(big.Rat).Set(m.amount)
}

func (m *Money) Numerator() int64 {
	return m.Amount().Num().Int64()
}

func (m *Money) Denominator() int64 {
	return m.Amount().Denom().Int64()
}

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

func (m *Money) String() string {
	return m.Amount().FloatString(2)
}
