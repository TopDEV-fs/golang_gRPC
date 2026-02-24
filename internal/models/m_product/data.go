// Package m_product provides the Spanner data struct and column-name constants
// for the products table. It is intentionally free of domain logic.
package m_product

import "time"

// Data is the raw Spanner row representation of a product. It is used only
// for scanning; domain objects are constructed via domain.RehydrateProduct.
type Data struct {
	ProductID            string
	Name                 string
	Description          string
	Category             string
	BasePriceNumerator   int64
	BasePriceDenominator int64
	DiscountPercent      string
	DiscountStartDate    time.Time
	DiscountEndDate      time.Time
	Status               string
	CreatedAt            time.Time
	UpdatedAt            time.Time
	ArchivedAt           time.Time
}
