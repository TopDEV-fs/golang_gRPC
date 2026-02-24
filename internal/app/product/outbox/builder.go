// Package outbox provides helpers for building transactional outbox event mutations.
//
// All mutations returned by this package are designed to be committed in the same
// Spanner read-write transaction as the aggregate mutation, guaranteeing at-least-once
// delivery semantics for downstream consumers.
package outbox

import (
	"encoding/json"
	"fmt"
	"time"

	"cloud.google.com/go/spanner"
	"github.com/google/uuid"

	"github.com/example/product-catalog-service/internal/app/product/contracts"
	"github.com/example/product-catalog-service/internal/app/product/domain"
)

// BuildMuts converts a slice of domain events into Spanner mutations for the
// outbox_events table. Each event is serialised to JSON and wrapped in a
// contracts.OutboxEvent with status PENDING.
//
// Returns an error only when JSON marshalling fails, which should never happen
// for the simple payload shape used here.
func BuildMuts(repo contracts.OutboxRepository, events []domain.DomainEvent, now time.Time) ([]*spanner.Mutation, error) {
	muts := make([]*spanner.Mutation, 0, len(events))
	for _, e := range events {
		payload, err := json.Marshal(map[string]any{
			"aggregate_id": e.AggregateID(),
			"event_type":   e.EventType(),
			"occurred_at":  e.OccurredAt().Format(time.RFC3339Nano),
		})
		if err != nil {
			return nil, fmt.Errorf("marshal outbox event %q: %w", e.EventType(), err)
		}
		mut := repo.InsertMut(contracts.OutboxEvent{
			EventID:      uuid.NewString(),
			EventType:    e.EventType(),
			AggregateID:  e.AggregateID(),
			Payload:      string(payload),
			Status:       "PENDING",
			CreatedAtUTC: now.Unix(),
		})
		if mut != nil {
			muts = append(muts, mut)
		}
	}
	return muts, nil
}
