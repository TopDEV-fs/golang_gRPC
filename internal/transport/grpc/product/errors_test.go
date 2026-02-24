package product

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/example/product-catalog-service/internal/app/product/domain"
)

func TestMapDomainError_InvalidArgument(t *testing.T) {
	invalidArgErrors := []error{
		domain.ErrInvalidName,
		domain.ErrInvalidCategory,
		domain.ErrInvalidPrice,
		domain.ErrInvalidDiscountPercent,
		domain.ErrInvalidDiscountPeriod,
	}
	for _, domainErr := range invalidArgErrors {
		got := mapDomainError(domainErr)
		st, ok := status.FromError(got)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code(), "expected InvalidArgument for %v", domainErr)
	}
}

func TestMapDomainError_FailedPrecondition(t *testing.T) {
	preconditionErrors := []error{
		domain.ErrProductNotActive,
		domain.ErrOverlappingDiscount,
		domain.ErrNoDiscount,
		domain.ErrProductArchived,
	}
	for _, domainErr := range preconditionErrors {
		got := mapDomainError(domainErr)
		st, ok := status.FromError(got)
		assert.True(t, ok)
		assert.Equal(t, codes.FailedPrecondition, st.Code(), "expected FailedPrecondition for %v", domainErr)
	}
}

func TestMapDomainError_UnknownBecomesInternal(t *testing.T) {
	got := mapDomainError(assert.AnError)
	st, ok := status.FromError(got)
	assert.True(t, ok)
	assert.Equal(t, codes.Internal, st.Code())
}
