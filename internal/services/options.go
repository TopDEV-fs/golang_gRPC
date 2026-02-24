// Package services wires together all application-layer components into a single
// Container that is consumed by the transport layer. It is the composition root
// for the product-catalog service.
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

// ProductCommands groups all write-side (command) use-case interactors.
type ProductCommands struct {
	CreateProduct *create_product.Interactor
	UpdateProduct *update_product.Interactor
	Activate      *activate_product.Interactor
	Discount      *apply_discount.Interactor
}

// ProductQueries groups all read-side (query) handlers.
type ProductQueries struct {
	GetProduct   *get_product.Query
	ListProducts *list_products.Query
}

// Container is the application-level service locator produced by BuildContainer.
// The transport layer MUST NOT instantiate use cases or queries directly;
// it should depend on Container exclusively.
type Container struct {
	Commands ProductCommands
	Queries  ProductQueries
}

// BuildContainer constructs and wires every application-layer component.
// It is called once at process start and the resulting Container is shared
// across all requests (all components are safe for concurrent use).
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
