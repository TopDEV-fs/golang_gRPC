package domain

import (
	"math/big"
	"time"
)

type Discount struct {
	percentage *big.Rat
	startDate  time.Time
	endDate    time.Time
}

func NewDiscount(percentage *big.Rat, startDate, endDate time.Time) (*Discount, error) {
	if percentage == nil {
		return nil, ErrInvalidDiscountPercent
	}
	if percentage.Sign() <= 0 || percentage.Cmp(big.NewRat(100, 1)) > 0 {
		return nil, ErrInvalidDiscountPercent
	}
	if endDate.Before(startDate) || endDate.Equal(startDate) {
		return nil, ErrInvalidDiscountPeriod
	}
	return &Discount{
		percentage: new(big.Rat).Set(percentage),
		startDate:  startDate.UTC(),
		endDate:    endDate.UTC(),
	}, nil
}

func (d *Discount) Percentage() *big.Rat {
	if d == nil || d.percentage == nil {
		return nil
	}
	return new(big.Rat).Set(d.percentage)
}

func (d *Discount) StartDate() time.Time { return d.startDate }
func (d *Discount) EndDate() time.Time   { return d.endDate }

func (d *Discount) IsValidAt(now time.Time) bool {
	if d == nil {
		return false
	}
	n := now.UTC()
	return (n.Equal(d.startDate) || n.After(d.startDate)) && n.Before(d.endDate)
}

func (d *Discount) Fraction() *big.Rat {
	return new(big.Rat).Quo(d.Percentage(), big.NewRat(100, 1))
}
