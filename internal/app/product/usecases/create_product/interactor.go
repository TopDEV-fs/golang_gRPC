package create_product

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Vektor-AI/commitplan"
	"github.com/google/uuid"
	"github.com/example/product-catalog-service/internal/app/product/contracts"
	"github.com/example/product-catalog-service/internal/app/product/domain"
	"github.com/example/product-catalog-service/internal/pkg/clock"
	pkgcommitter "github.com/example/product-catalog-service/internal/pkg/committer"
)

type Request struct {
	Name                 string
	Description          string
	Category             string
	BasePriceNumerator   int64
	BasePriceDenominator int64
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

	if err := it.committer.Apply(ctx, plan); err != nil {
		return "", err
	}
	return product.ID(), nil
}
