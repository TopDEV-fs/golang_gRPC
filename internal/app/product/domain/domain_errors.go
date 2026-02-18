package domain

import "errors"

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