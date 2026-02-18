package product

import (
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"github.com/example/product-catalog-service/internal/app/product/domain"
)

func mapDomainError(err error) error {
	switch {
	case errors.Is(err, domain.ErrInvalidName),
		errors.Is(err, domain.ErrInvalidCategory),
		errors.Is(err, domain.ErrInvalidPrice),
		errors.Is(err, domain.ErrInvalidDiscountPercent),
		errors.Is(err, domain.ErrInvalidDiscountPeriod):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, domain.ErrProductNotActive),
		errors.Is(err, domain.ErrOverlappingDiscount),
		errors.Is(err, domain.ErrNoDiscount),
		errors.Is(err, domain.ErrProductArchived):
		return status.Error(codes.FailedPrecondition, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}
