package productv1

type CreateProductRequest struct {
	Name                 string
	Description          string
	Category             string
	BasePriceNumerator   int64
	BasePriceDenominator int64
}

type CreateProductReply struct{ ProductId string }

type UpdateProductRequest struct {
	ProductId    string
	Name         string
	Description  string
	Category     string
}

type UpdateProductReply struct{}

type ActivateProductRequest struct{ ProductId string }
type ActivateProductReply struct{}

type DeactivateProductRequest struct{ ProductId string }
type DeactivateProductReply struct{}

type ApplyDiscountRequest struct {
	ProductId      string
	Percent        string
	StartDateUnix  int64
	EndDateUnix    int64
}

type ApplyDiscountReply struct{}

type RemoveDiscountRequest struct{ ProductId string }
type RemoveDiscountReply struct{}

type GetProductRequest struct{ ProductId string }

type GetProductReply struct {
	ProductId      string
	Name           string
	Description    string
	Category       string
	Status         string
	BasePrice      string
	EffectivePrice string
	Discount       *Discount
}

type ListProductsRequest struct {
	PageSize  int32
	PageToken string
	Category  string
}

type ListProductsReply struct {
	Products      []*Product
	NextPageToken string
}

type Product struct {
	ProductId      string
	Name           string
	Description    string
	Category       string
	Status         string
	BasePrice      string
	EffectivePrice string
}

type Discount struct {
	Percent       string
	StartDateUnix int64
	EndDateUnix   int64
}
