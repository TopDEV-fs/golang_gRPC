package services

import (
	"cloud.google.com/go/spanner"
	"github.com/example/product-catalog-service/internal/app/product/queries/get_product"
	"github.com/example/product-catalog-service/internal/app/product/queries/list_products"
	"github.com/example/product-catalog-service/internal/app/product/repo"
	"github.com/example/product-catalog-service/internal/app/product/usecases/activate_product"
	"github.com/example/product-catalog-service/internal/app/product/usecases/apply_discount"
	"github.com/example/product-catalog-service/internal/app/product/usecases/create_product"
	"github.com/example/product-catalog-service/internal/app/product/usecases/update_product"
	"github.com/example/product-catalog-service/internal/pkg/clock"
	"github.com/example/product-catalog-service/internal/pkg/committer"
)

type ProductCommands struct {
	CreateProduct *create_product.Interactor
	UpdateProduct *update_product.Interactor
	Activate      *activate_product.Interactor
	Discount      *apply_discount.Interactor
}

type ProductQueries struct {
	GetProduct   *get_product.Query
	ListProducts *list_products.Query
}

type Container struct {
	Commands ProductCommands
	Queries  ProductQueries
}

func BuildContainer(spannerClient *spanner.Client) *Container {
	productRepo := repo.NewProductRepo(spannerClient)
	outboxRepo := repo.NewOutboxRepo()
	readModel := repo.NewProductReadModel(spannerClient)
	clk := clock.NewRealClock()
	cp := committer.NewSpannerCommitter(spannerClient)

	return &Container{
		Commands: ProductCommands{
			CreateProduct: create_product.New(productRepo, outboxRepo, cp, clk),
			UpdateProduct: update_product.New(productRepo, outboxRepo, cp, clk),
			Activate:      activate_product.New(productRepo, outboxRepo, cp, clk),
			Discount:      apply_discount.New(productRepo, outboxRepo, cp, clk),
		},
		Queries: ProductQueries{
			GetProduct:   get_product.New(readModel),
			ListProducts: list_products.New(readModel),
		},
	}
}
