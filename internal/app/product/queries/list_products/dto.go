package list_products

type ProductDTO struct {
	ID             string
	Name           string
	Description    string
	Category       string
	Status         string
	BasePrice      string
	EffectivePrice string
}

type Result struct {
	Items         []ProductDTO
	NextPageToken string
}
