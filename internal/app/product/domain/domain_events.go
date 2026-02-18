package domain

import "time"

type DomainEvent interface {
	EventType() string
	AggregateID() string
	OccurredAt() time.Time
}

type baseEvent struct {
	aggregateID string
	occurredAt time.Time
}

func (e baseEvent) AggregateID() string { return e.aggregateID }
func (e baseEvent) OccurredAt() time.Time {
	return e.occurredAt
}

type ProductCreatedEvent struct{ baseEvent }
type ProductUpdatedEvent struct{ baseEvent }
type ProductActivatedEvent struct{ baseEvent }
type ProductDeactivatedEvent struct{ baseEvent }
type DiscountAppliedEvent struct{ baseEvent }
type DiscountRemovedEvent struct{ baseEvent }

func (ProductCreatedEvent) EventType() string    { return "product.created" }
func (ProductUpdatedEvent) EventType() string    { return "product.updated" }
func (ProductActivatedEvent) EventType() string  { return "product.activated" }
func (ProductDeactivatedEvent) EventType() string { return "product.deactivated" }
func (DiscountAppliedEvent) EventType() string   { return "product.discount_applied" }
func (DiscountRemovedEvent) EventType() string   { return "product.discount_removed" }
