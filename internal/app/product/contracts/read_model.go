package contracts

import "context"

// ProductReadModel is the read-side query interface used by the query layer.
// Implementations perform direct SQL reads and return flat DTOs, bypassing
// the domain aggregate entirely in the spirit of CQRS.
type ProductReadModel interface {
	// GetByID returns a single product DTO, or an error (iterator.Done when missing).
	GetByID(ctx context.Context, id string) (*ProductDTO, error)
	// ListActive returns a paginated list of ACTIVE products, optionally filtered
	// by category. pageToken is an opaque offset token; pass empty for the first page.
	ListActive(ctx context.Context, category string, pageSize int32, pageToken string) (*ListProductsResult, error)
}

// ProductDTO is the flat, read-model representation of a product as stored in
// Spanner. Numeric fields use their raw int64/string forms to preserve precision
// across serialisation boundaries.
type ProductDTO struct {
	ID                   string
	Name                 string
	Description          string
	Category             string
	Status               string
	BasePriceNumerator   int64
	BasePriceDenominator int64
	DiscountPercent      string
	DiscountStartUnix    int64
	DiscountEndUnix      int64
	CreatedAtUnix        int64
	UpdatedAtUnix        int64
	ArchivedAtUnix       int64
}

// ListProductsResult wraps a page of ProductDTO items and the token for the next page.
type ListProductsResult struct {
	Items         []ProductDTO
	NextPageToken string
}
