package update_product

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Vektor-AI/commitplan"
	"github.com/google/uuid"
	"github.com/example/product-catalog-service/internal/app/product/contracts"
	"github.com/example/product-catalog-service/internal/pkg/clock"
	pkgcommitter "github.com/example/product-catalog-service/internal/pkg/committer"
)

type Request struct {
	ProductID    string
	Name         string
	Description  string
	Category     string
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
