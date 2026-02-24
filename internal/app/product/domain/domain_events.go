package domain

import "time"

// DomainEvent is the common interface implemented by all product domain events.
// Events are raised by aggregate business operations and consumed by use cases
// to populate the transactional outbox for downstream processing.
type DomainEvent interface {
	// EventType returns a stable, dot-separated type identifier (e.g. "product.created").
	EventType() string
	// AggregateID returns the identifier of the aggregate that raised the event.
	AggregateID() string
	// OccurredAt returns the UTC instant at which the event was raised.
	OccurredAt() time.Time
}

// baseEvent holds fields shared by all concrete product event types.
type baseEvent struct {
	aggregateID string
	occurredAt  time.Time
}

func (e baseEvent) AggregateID() string   { return e.aggregateID }
func (e baseEvent) OccurredAt() time.Time { return e.occurredAt }

// ProductCreatedEvent is raised when a new product is first persisted.
type ProductCreatedEvent struct{ baseEvent }

// ProductUpdatedEvent is raised when product details (name/description/category) change.
type ProductUpdatedEvent struct{ baseEvent }

// ProductActivatedEvent is raised when a product transitions to the ACTIVE status.
type ProductActivatedEvent struct{ baseEvent }

// ProductDeactivatedEvent is raised when a product transitions to the INACTIVE status.
type ProductDeactivatedEvent struct{ baseEvent }

// DiscountAppliedEvent is raised when a new discount is successfully applied.
type DiscountAppliedEvent struct{ baseEvent }

// DiscountRemovedEvent is raised when an existing discount is removed.
type DiscountRemovedEvent struct{ baseEvent }

func (ProductCreatedEvent) EventType() string     { return "product.created" }
func (ProductUpdatedEvent) EventType() string     { return "product.updated" }
func (ProductActivatedEvent) EventType() string   { return "product.activated" }
func (ProductDeactivatedEvent) EventType() string { return "product.deactivated" }
func (DiscountAppliedEvent) EventType() string    { return "product.discount_applied" }
func (DiscountRemovedEvent) EventType() string    { return "product.discount_removed" }
