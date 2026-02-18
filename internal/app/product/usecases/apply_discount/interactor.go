package apply_discount

import (
	"context"
	"encoding/json"
	"math/big"
	"time"

	"github.com/Vektor-AI/commitplan"
	"github.com/google/uuid"
	"github.com/example/product-catalog-service/internal/app/product/contracts"
	"github.com/example/product-catalog-service/internal/app/product/domain"
	"github.com/example/product-catalog-service/internal/pkg/clock"
	pkgcommitter "github.com/example/product-catalog-service/internal/pkg/committer"
)

type ApplyRequest struct {
	ProductID    string
	Percent      string
	StartDateUTC time.Time
	EndDateUTC   time.Time
}

type RemoveRequest struct {
	ProductID string
}

type Interactor struct {
	repo      contracts.ProductRepository
	outbox    contracts.OutboxRepository
	committer pkgcommitter.PlanApplier
	clock     clock.Clock
}

func New(repo contracts.ProductRepository, outbox contracts.OutboxRepository, committer pkgcommitter.PlanApplier, clock clock.Clock) *Interactor {
	return &Interactor{repo: repo, outbox: outbox, committer: committer, clock: clock}
}

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

func (it *Interactor) applyPlan(ctx context.Context, product *domain.Product, now time.Time) error {
	plan := commitplan.NewPlan()
	if mut := it.repo.UpdateMut(product); mut != nil {
		plan.Add(mut)
	}
	for _, event := range product.DomainEvents() {
		payload, _ := json.Marshal(map[string]any{
			"aggregate_id": event.AggregateID(),
			"event_type":   event.EventType(),
			"occurred_at":  event.OccurredAt().Format(time.RFC3339Nano),
		})
		if mut := it.outbox.InsertMut(contracts.OutboxEvent{
			EventID:      uuid.NewString(),
			EventType:    event.EventType(),
			AggregateID:  event.AggregateID(),
			Payload:      string(payload),
			Status:       "PENDING",
			CreatedAtUTC: now.Unix(),
		}); mut != nil {
			plan.Add(mut)
		}
	}
	return it.committer.Apply(ctx, plan)
}
