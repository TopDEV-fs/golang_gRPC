// Package apply_discount implements the ApplyDiscount and RemoveDiscount command use cases.
//
// Both commands enforce business rules (product must be active, discount period must be
// valid and non-overlapping) before persisting the aggregate change and outbox event
// in an atomic transaction.
package apply_discount

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/Vektor-AI/commitplan"

	"github.com/example/product-catalog-service/internal/app/product/contracts"
	"github.com/example/product-catalog-service/internal/app/product/domain"
	"github.com/example/product-catalog-service/internal/app/product/outbox"
	"github.com/example/product-catalog-service/internal/pkg/clock"
	pkgcommitter "github.com/example/product-catalog-service/internal/pkg/committer"
)

// ApplyRequest carries the input for applying a percentage discount to a product.
type ApplyRequest struct {
	ProductID    string
	Percent      string // decimal string, e.g. "20" or "12.5"
	StartDateUTC time.Time
	EndDateUTC   time.Time
}

// RemoveRequest carries the product identifier for the RemoveDiscount command.
type RemoveRequest struct {
	ProductID string
}

// Interactor executes the apply-discount and remove-discount commands.
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

// Apply parses the discount percentage, validates business rules, and commits the
// updated aggregate together with the outbox event in one transaction.
func (it *Interactor) Apply(ctx context.Context, req ApplyRequest) error {
	product, err := it.repo.FindByID(ctx, req.ProductID)
	if err != nil {
		return err
	}
	percent, ok := new(big.Rat).SetString(req.Percent)
	if !ok {
		return domain.ErrInvalidDiscountPercent
	}
	discount, err := domain.NewDiscount(percent, req.StartDateUTC, req.EndDateUTC)
	if err != nil {
		return err
	}
	now := it.clock.Now()
	if err := product.ApplyDiscount(discount, now); err != nil {
		return err
	}
	return it.applyPlan(ctx, product, now)
}

// Remove clears the active discount from the product and commits the change atomically.
func (it *Interactor) Remove(ctx context.Context, req RemoveRequest) error {
	product, err := it.repo.FindByID(ctx, req.ProductID)
	if err != nil {
		return err
	}
	now := it.clock.Now()
	if err := product.RemoveDiscount(now); err != nil {
		return err
	}
	return it.applyPlan(ctx, product, now)
}

// applyPlan builds and applies the commit plan containing aggregate and outbox mutations.
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
