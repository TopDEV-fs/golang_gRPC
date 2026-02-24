// Package create_product implements the CreateProduct command use case.
//
// It validates input, constructs the Product aggregate, and atomically persists
// both the aggregate row and the corresponding transactional outbox event.
package create_product

import (
	"context"
	"fmt"

	"github.com/Vektor-AI/commitplan"
	"github.com/google/uuid"

	"github.com/example/product-catalog-service/internal/app/product/contracts"
	"github.com/example/product-catalog-service/internal/app/product/domain"
	"github.com/example/product-catalog-service/internal/app/product/outbox"
	"github.com/example/product-catalog-service/internal/pkg/clock"
	pkgcommitter "github.com/example/product-catalog-service/internal/pkg/committer"
)

// Request carries the input data for the CreateProduct command.
type Request struct {
	Name                 string
	Description          string
	Category             string
	BasePriceNumerator   int64
	BasePriceDenominator int64
}

// Interactor orchestrates the CreateProduct command: it constructs the aggregate,
// builds a commit plan containing the product insert and outbox event mutations,
// and applies the plan atomically.
type Interactor struct {
	repo      contracts.ProductRepository
	outbox    contracts.OutboxRepository
	committer pkgcommitter.PlanApplier
	clock     clock.Clock
}

// New returns a new CreateProduct Interactor wired with its dependencies.
func New(repo contracts.ProductRepository, outbox contracts.OutboxRepository, committer pkgcommitter.PlanApplier, clock clock.Clock) *Interactor {
	return &Interactor{repo: repo, outbox: outbox, committer: committer, clock: clock}
}

// Execute creates a new product and persists it together with its domain events
// in a single atomic transaction. Returns the new product ID on success.
func (it *Interactor) Execute(ctx context.Context, req Request) (string, error) {
	basePrice, err := domain.NewMoney(req.BasePriceNumerator, req.BasePriceDenominator)
	if err != nil {
		return "", err
	}

	now := it.clock.Now()
	product, err := domain.NewProduct(uuid.NewString(), req.Name, req.Description, req.Category, basePrice, now)
	if err != nil {
		return "", err
	}

	plan := commitplan.NewPlan()
	if mut := it.repo.InsertMut(product); mut != nil {
		plan.Add(mut)
	}

	outboxMuts, err := outbox.BuildMuts(it.outbox, product.PullDomainEvents(), now)
	if err != nil {
		return "", fmt.Errorf("build outbox mutations: %w", err)
	}
	for _, m := range outboxMuts {
		plan.Add(m)
	}

	if err := it.committer.Apply(ctx, plan); err != nil {
		return "", err
	}
	return product.ID(), nil
}
