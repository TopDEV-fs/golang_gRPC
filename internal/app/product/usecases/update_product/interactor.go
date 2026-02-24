// Package update_product implements the UpdateProduct command use case.
//
// It loads the product aggregate, applies field-level updates (name, description,
// category), and atomically persists the changed fields alongside the outbox event.
package update_product

import (
	"context"
	"fmt"

	"github.com/Vektor-AI/commitplan"

	"github.com/example/product-catalog-service/internal/app/product/contracts"
	"github.com/example/product-catalog-service/internal/app/product/outbox"
	"github.com/example/product-catalog-service/internal/pkg/clock"
	pkgcommitter "github.com/example/product-catalog-service/internal/pkg/committer"
)

// Request carries the input fields for the UpdateProduct command.
// Only non-empty fields are applied; empty strings are treated as no-op by the aggregate.
type Request struct {
	ProductID   string
	Name        string
	Description string
	Category    string
}

// Interactor executes the UpdateProduct command.
type Interactor struct {
	repo      contracts.ProductRepository
	outbox    contracts.OutboxRepository
	committer pkgcommitter.PlanApplier
	clock     clock.Clock
}

// New returns a new UpdateProduct Interactor wired with its dependencies.
func New(repo contracts.ProductRepository, outbox contracts.OutboxRepository, committer pkgcommitter.PlanApplier, clock clock.Clock) *Interactor {
	return &Interactor{repo: repo, outbox: outbox, committer: committer, clock: clock}
}

// Execute loads the product, mutates it according to the request, and persists
// only the dirty fields in a targeted update mutation alongside any outbox events.
func (it *Interactor) Execute(ctx context.Context, req Request) error {
	product, err := it.repo.FindByID(ctx, req.ProductID)
	if err != nil {
		return err
	}
	now := it.clock.Now()
	if err := product.UpdateDetails(req.Name, req.Description, req.Category, now); err != nil {
		return err
	}

	plan := commitplan.NewPlan()
	if mut := it.repo.UpdateMut(product); mut != nil {
		plan.Add(mut)
	}
	outboxMuts, err := outbox.BuildMuts(it.outbox, product.PullDomainEvents(), now)
	if err != nil {
		return fmt.Errorf("build outbox mutations: %w", err)
	}
	for _, m := range outboxMuts {
		plan.Add(m)
	}
	return it.committer.Apply(ctx, plan)
}
