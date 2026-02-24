// Package get_product implements the GetProduct query, resolving the effective
// price (with active discount if present) for a single product.
package get_product

// ProductDTO is the query-layer response for a single product lookup.
// BasePrice and EffectivePrice are pre-formatted decimal strings (e.g. "19.99").
type ProductDTO struct {
	ID             string
	Name           string
	Description    string
	Category       string
	Status         string
	BasePrice      string
	EffectivePrice string
	Discount       *DiscountDTO
	CreatedAtUnix  int64
	UpdatedAtUnix  int64
}

// DiscountDTO carries the active discount details embedded in a ProductDTO response.
type DiscountDTO struct {
	Percent       string
	StartDateUnix int64
	EndDateUnix   int64
}
