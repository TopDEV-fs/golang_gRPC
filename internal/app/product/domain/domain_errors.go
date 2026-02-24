// Package domain contains the Product bounded-context aggregate, value objects,
// domain events, and pure business rules. It has no dependencies on persistence,
// transport, or any external framework.
package domain

import "errors"

// Sentinel errors returned by domain operations. Callers should use errors.Is to
// test for these values; never compare by string message.
var (
	ErrInvalidName            = errors.New("invalid product name")
	ErrInvalidCategory        = errors.New("invalid product category")
	ErrInvalidPrice           = errors.New("invalid base price")
	ErrProductNotActive       = errors.New("product is not active")
	ErrProductArchived        = errors.New("product is archived")
	ErrInvalidDiscountPercent = errors.New("invalid discount percent")
	ErrInvalidDiscountPeriod  = errors.New("invalid discount period")
	ErrOverlappingDiscount    = errors.New("product already has an active discount")
	ErrNoDiscount             = errors.New("product has no active discount")
)