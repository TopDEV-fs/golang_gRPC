// Package contracts defines the repository and read-model interfaces required
// by the product application layer. Implementations live in the repo package;
// tests may provide fakes.
package contracts

import (
	"context"

	"cloud.google.com/go/spanner"
	"github.com/example/product-catalog-service/internal/app/product/domain"
)

// ProductRepository is the write-side repository contract for the Product aggregate.
// Methods that return spanner.Mutation are designed for the Golden Mutation Pattern:
// the caller adds every mutation to a commit plan and applies it atomically.
type ProductRepository interface {
	// FindByID loads and rehydrates a Product aggregate by its identifier.
	// Returns iterator.Done (wrapped) when the product does not exist.
	FindByID(ctx context.Context, id string) (*domain.Product, error)
	// InsertMut returns a Spanner insert mutation for a newly created product.
	InsertMut(product *domain.Product) *spanner.Mutation
	// UpdateMut returns a targeted Spanner update mutation for the dirty fields
	// of an existing product. Returns nil when no fields are dirty.
	UpdateMut(product *domain.Product) *spanner.Mutation
}

// OutboxRepository is the write-side repository contract for the outbox_events table.
type OutboxRepository interface {
	// InsertMut returns a Spanner insert mutation for the given outbox event.
	InsertMut(event OutboxEvent) *spanner.Mutation
}

// OutboxEvent is the persistence representation of a domain event destined for
// the transactional outbox. All fields are scalar to remain serialisation-agnostic.
type OutboxEvent struct {
	EventID      string
	EventType    string
	AggregateID  string
	Payload      string
	Status       string
	CreatedAtUTC int64
}
