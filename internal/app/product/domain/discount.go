package domain

import (
	"math/big"
	"time"
)

// Discount is an immutable value object representing a percentage-based price
// reduction that is valid within a specific UTC date range. The percentage is
// stored as a *big.Rat to preserve precision across arithmetic operations.
type Discount struct {
	percentage *big.Rat
	startDate  time.Time
	endDate    time.Time
}

// NewDiscount constructs a Discount and validates:
//   - percentage is strictly between 0 and 100 (inclusive of neither boundary)
//   - endDate is strictly after startDate
//
// Returns ErrInvalidDiscountPercent or ErrInvalidDiscountPeriod on violation.
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

// Percentage returns a copy of the discount percentage as a *big.Rat (e.g. 20 for 20%).
func (d *Discount) Percentage() *big.Rat {
	if d == nil || d.percentage == nil {
		return nil
	}
	return new(big.Rat).Set(d.percentage)
}

// StartDate returns the UTC start of the discount validity window.
func (d *Discount) StartDate() time.Time { return d.startDate }

// EndDate returns the UTC end of the discount validity window (exclusive).
func (d *Discount) EndDate() time.Time { return d.endDate }

// IsValidAt reports whether the discount is active at the given instant.
func (d *Discount) IsValidAt(now time.Time) bool {
	if d == nil {
		return false
	}
	n := now.UTC()
	return (n.Equal(d.startDate) || n.After(d.startDate)) && n.Before(d.endDate)
}

// Fraction returns the discount as a decimal fraction suitable for multiplication
// (e.g. 20% → 0.20 expressed as a *big.Rat).
func (d *Discount) Fraction() *big.Rat {
	return new(big.Rat).Quo(d.Percentage(), big.NewRat(100, 1))
}
