package outbox_test

import (
	"testing"
	"time"

	"cloud.google.com/go/spanner"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/example/product-catalog-service/internal/app/product/contracts"
	"github.com/example/product-catalog-service/internal/app/product/domain"
	"github.com/example/product-catalog-service/internal/app/product/outbox"
)

// fakeOutboxRepo captures InsertMut calls without touching Spanner.
type fakeOutboxRepo struct {
	captured []contracts.OutboxEvent
}

func (r *fakeOutboxRepo) InsertMut(e contracts.OutboxEvent) *spanner.Mutation {
	r.captured = append(r.captured, e)
	return nil
}

type fakeEvent struct {
	eventType   string
	aggregateID string
	occurredAt  time.Time
}

func (e fakeEvent) EventType() string   { return e.eventType }
func (e fakeEvent) AggregateID() string { return e.aggregateID }
func (e fakeEvent) OccurredAt() time.Time { return e.occurredAt }

func TestBuildMuts_EmptyEvents(t *testing.T) {
	repo := &fakeOutboxRepo{}
	muts, err := outbox.BuildMuts(repo, nil, time.Now())
	require.NoError(t, err)
	assert.Empty(t, muts)
	assert.Empty(t, repo.captured)
}

func TestBuildMuts_PopulatesOutboxEvent(t *testing.T) {
	repo := &fakeOutboxRepo{}
	now := time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC)

	events := []domain.DomainEvent{
		fakeEvent{eventType: "product.created", aggregateID: "agg-1", occurredAt: now},
		fakeEvent{eventType: "product.activated", aggregateID: "agg-1", occurredAt: now},
	}

	_, err := outbox.BuildMuts(repo, events, now)
	require.NoError(t, err)

	require.Len(t, repo.captured, 2)

	first := repo.captured[0]
	assert.Equal(t, "product.created", first.EventType)
	assert.Equal(t, "agg-1", first.AggregateID)
	assert.Equal(t, "PENDING", first.Status)
	assert.NotEmpty(t, first.EventID, "UUID must be generated")
	assert.Equal(t, now.Unix(), first.CreatedAtUTC)

	second := repo.captured[1]
	assert.Equal(t, "product.activated", second.EventType)
	// EventIDs must be unique per event.
	assert.NotEqual(t, first.EventID, second.EventID)
}

// Compile-time check: fakeOutboxRepo satisfies the contract.
var _ contracts.OutboxRepository = (*fakeOutboxRepo)(nil)
