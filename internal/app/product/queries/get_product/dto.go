package get_product

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

type DiscountDTO struct {
	Percent       string
	StartDateUnix int64
	EndDateUnix   int64
}
