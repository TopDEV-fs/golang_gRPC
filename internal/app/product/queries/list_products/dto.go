// Package list_products implements the ListProducts query, returning paginated
// active products with effective prices calculated at query time.
package list_products

// ProductDTO is the lightweight read-model representation returned per product
// in a listing response.
type ProductDTO struct {
	ID             string
	Name           string
	Description    string
	Category       string
	Status         string
	BasePrice      string
	EffectivePrice string
}

// Result wraps a page of ProductDTO items and the pagination token for subsequent requests.
type Result struct {
	Items         []ProductDTO
	NextPageToken string
}
