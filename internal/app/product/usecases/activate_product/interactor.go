// Package activate_product implements the Activate/Deactivate Product command use cases.
//
// A single Interactor handles both state transitions; each method loads the aggregate,
// applies the domain transition, and commits the result plus domain events atomically.
package activate_product

import (
	"context"
	"fmt"
	"time"

	"github.com/Vektor-AI/commitplan"

	"github.com/example/product-catalog-service/internal/app/product/contracts"
	"github.com/example/product-catalog-service/internal/app/product/domain"
	"github.com/example/product-catalog-service/internal/app/product/outbox"
	"github.com/example/product-catalog-service/internal/pkg/clock"
	pkgcommitter "github.com/example/product-catalog-service/internal/pkg/committer"
)

// ActivateRequest is the input for the Activate command.
type ActivateRequest struct {
	ProductID string
}

// DeactivateRequest is the input for the Deactivate command.
type DeactivateRequest struct {
	ProductID string
}

// Interactor executes the activate and deactivate product commands.
type Interactor struct {
	repo      contracts.ProductRepository
	outbox    contracts.OutboxRepository
	committer pkgcommitter.PlanApplier
	clock     clock.Clock
}

// New returns a new Interactor wired with its dependencies.
func New(repo contracts.ProductRepository, outbox contracts.OutboxRepository, committer pkgcommitter.PlanApplier, clock clock.Clock) *Interactor {
	return &Interactor{repo: repo, outbox: outbox, committer: committer, clock: clock}
}

// Activate transitions a product to the ACTIVE status. Returns an error if the
// product is already archived or the transition would otherwise violate business rules.
func (it *Interactor) Activate(ctx context.Context, req ActivateRequest) error {
	product, err := it.repo.FindByID(ctx, req.ProductID)
	if err != nil {
		return err
	}
	now := it.clock.Now()
	if err := product.Activate(now); err != nil {
		return err
	}
	return it.applyPlan(ctx, product, now)
}

// Deactivate transitions a product to the INACTIVE status. Returns an error if the
// product is already archived or the transition would otherwise violate business rules.
func (it *Interactor) Deactivate(ctx context.Context, req DeactivateRequest) error {
	product, err := it.repo.FindByID(ctx, req.ProductID)
	if err != nil {
		return err
	}
	now := it.clock.Now()
	if err := product.Deactivate(now); err != nil {
		return err
	}
	return it.applyPlan(ctx, product, now)
}

// applyPlan builds a commit plan containing the aggregate update and outbox mutations,
// then applies the plan atomically through the PlanApplier.
func (it *Interactor) applyPlan(ctx context.Context, product *domain.Product, now time.Time) error {
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
