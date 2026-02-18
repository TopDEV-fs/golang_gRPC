package contracts

import "context"

type ProductReadModel interface {
	GetByID(ctx context.Context, id string) (*ProductDTO, error)
	ListActive(ctx context.Context, category string, pageSize int32, pageToken string) (*ListProductsResult, error)
}

type ProductDTO struct {
	ID                    string
	Name                  string
	Description           string
	Category              string
	Status                string
	BasePriceNumerator    int64
	BasePriceDenominator  int64
	DiscountPercent       string
	DiscountStartUnix     int64
	DiscountEndUnix       int64
	CreatedAtUnix         int64
	UpdatedAtUnix         int64
	ArchivedAtUnix        int64
}

type ListProductsResult struct {
	Items         []ProductDTO
	NextPageToken string
}
