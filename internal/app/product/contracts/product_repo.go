package contracts

import (
	"context"

	"cloud.google.com/go/spanner"
	"github.com/example/product-catalog-service/internal/app/product/domain"
)

type ProductRepository interface {
	FindByID(ctx context.Context, id string) (*domain.Product, error)
	InsertMut(product *domain.Product) *spanner.Mutation
	UpdateMut(product *domain.Product) *spanner.Mutation
}

type OutboxRepository interface {
	InsertMut(event OutboxEvent) *spanner.Mutation
}

type OutboxEvent struct {
	EventID      string
	EventType    string
	AggregateID  string
	Payload      string
	Status       string
	CreatedAtUTC int64
}
