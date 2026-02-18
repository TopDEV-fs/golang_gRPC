package domain

import (
	"strings"
	"time"
)

const (
	FieldName        = "name"
	FieldDescription = "description"
	FieldCategory    = "category"
	FieldStatus      = "status"
	FieldDiscount    = "discount"
	FieldArchivedAt  = "archived_at"
)

type ProductStatus string

const (
	ProductStatusInactive ProductStatus = "INACTIVE"
	ProductStatusActive   ProductStatus = "ACTIVE"
	ProductStatusArchived ProductStatus = "ARCHIVED"
)

type ChangeTracker struct {
	dirtyFields map[string]bool
}

func NewChangeTracker() *ChangeTracker {
	return &ChangeTracker{dirtyFields: make(map[string]bool)}
}

func (ct *ChangeTracker) MarkDirty(field string) { ct.dirtyFields[field] = true }
func (ct *ChangeTracker) Dirty(field string) bool { return ct.dirtyFields[field] }
func (ct *ChangeTracker) Fields() map[string]bool {
	copy := make(map[string]bool, len(ct.dirtyFields))
	for k, v := range ct.dirtyFields {
		copy[k] = v
	}
	return copy
}
func (ct *ChangeTracker) Reset() { ct.dirtyFields = make(map[string]bool) }

type Product struct {
	id          string
	name        string
	description string
	category    string
	basePrice   *Money
	discount    *Discount
	status      ProductStatus
	createdAt   time.Time
	updatedAt   time.Time
	archivedAt  *time.Time
	changes     *ChangeTracker
	events      []DomainEvent
}

func NewProduct(id, name, description, category string, basePrice *Money, now time.Time) (*Product, error) {
	if strings.TrimSpace(name) == "" {
		return nil, ErrInvalidName
	}
	if strings.TrimSpace(category) == "" {
		return nil, ErrInvalidCategory
	}
	if basePrice == nil || basePrice.Amount().Sign() <= 0 {
		return nil, ErrInvalidPrice
	}
	n := now.UTC()
	p := &Product{
		id:          id,
		name:        strings.TrimSpace(name),
		description: strings.TrimSpace(description),
		category:    strings.TrimSpace(category),
		basePrice:   basePrice,
		status:      ProductStatusInactive,
		createdAt:   n,
		updatedAt:   n,
		changes:     NewChangeTracker(),
	}
	p.events = append(p.events, ProductCreatedEvent{baseEvent{aggregateID: p.id, occurredAt: n}})
	return p, nil
}

func RehydrateProduct(id, name, description, category string, basePrice *Money, discount *Discount, status ProductStatus, createdAt, updatedAt time.Time, archivedAt *time.Time) *Product {
	return &Product{
		id:          id,
		name:        name,
		description: description,
		category:    category,
		basePrice:   basePrice,
		discount:    discount,
		status:      status,
		createdAt:   createdAt.UTC(),
		updatedAt:   updatedAt.UTC(),
		archivedAt:  archivedAt,
		changes:     NewChangeTracker(),
		events:      nil,
	}
}

func (p *Product) UpdateDetails(name, description, category string, now time.Time) error {
	if p.status == ProductStatusArchived {
		return ErrProductArchived
	}
	trimmedName := strings.TrimSpace(name)
	trimmedCategory := strings.TrimSpace(category)
	if trimmedName == "" {
		return ErrInvalidName
	}
	if trimmedCategory == "" {
		return ErrInvalidCategory
	}
	if p.name != trimmedName {
		p.name = trimmedName
		p.changes.MarkDirty(FieldName)
	}
	if p.description != strings.TrimSpace(description) {
		p.description = strings.TrimSpace(description)
		p.changes.MarkDirty(FieldDescription)
	}
	if p.category != trimmedCategory {
		p.category = trimmedCategory
		p.changes.MarkDirty(FieldCategory)
	}
	p.touch(now)
	if len(p.changes.Fields()) > 0 {
		p.events = append(p.events, ProductUpdatedEvent{baseEvent{aggregateID: p.id, occurredAt: now.UTC()}})
	}
	return nil
}

func (p *Product) Activate(now time.Time) error {
	if p.status == ProductStatusArchived {
		return ErrProductArchived
	}
	if p.status != ProductStatusActive {
		p.status = ProductStatusActive
		p.changes.MarkDirty(FieldStatus)
		p.touch(now)
		p.events = append(p.events, ProductActivatedEvent{baseEvent{aggregateID: p.id, occurredAt: now.UTC()}})
	}
	return nil
}

func (p *Product) Deactivate(now time.Time) error {
	if p.status == ProductStatusArchived {
		return ErrProductArchived
	}
	if p.status != ProductStatusInactive {
		p.status = ProductStatusInactive
		p.changes.MarkDirty(FieldStatus)
		p.touch(now)
		p.events = append(p.events, ProductDeactivatedEvent{baseEvent{aggregateID: p.id, occurredAt: now.UTC()}})
	}
	return nil
}

func (p *Product) Archive(now time.Time) error {
	if p.status == ProductStatusArchived {
		return nil
	}
	n := now.UTC()
	p.status = ProductStatusArchived
	p.archivedAt = &n
	p.changes.MarkDirty(FieldStatus)
	p.changes.MarkDirty(FieldArchivedAt)
	p.touch(now)
	return nil
}

func (p *Product) ApplyDiscount(discount *Discount, now time.Time) error {
	if p.status != ProductStatusActive {
		return ErrProductNotActive
	}
	if !discount.IsValidAt(now) {
		return ErrInvalidDiscountPeriod
	}
	if p.discount != nil && p.discount.IsValidAt(now) {
		return ErrOverlappingDiscount
	}
	p.discount = discount
	p.changes.MarkDirty(FieldDiscount)
	p.touch(now)
	p.events = append(p.events, DiscountAppliedEvent{baseEvent{aggregateID: p.id, occurredAt: now.UTC()}})
	return nil
}

func (p *Product) RemoveDiscount(now time.Time) error {
	if p.discount == nil {
		return ErrNoDiscount
	}
	p.discount = nil
	p.changes.MarkDirty(FieldDiscount)
	p.touch(now)
	p.events = append(p.events, DiscountRemovedEvent{baseEvent{aggregateID: p.id, occurredAt: now.UTC()}})
	return nil
}

func (p *Product) touch(now time.Time) {
	p.updatedAt = now.UTC()
}

func (p *Product) PullDomainEvents() []DomainEvent {
	e := p.events
	p.events = nil
	return e
}

func (p *Product) DomainEvents() []DomainEvent { return p.events }
func (p *Product) Changes() *ChangeTracker      { return p.changes }
func (p *Product) ID() string                   { return p.id }
func (p *Product) Name() string                 { return p.name }
func (p *Product) Description() string          { return p.description }
func (p *Product) Category() string             { return p.category }
func (p *Product) BasePrice() *Money            { return p.basePrice }
func (p *Product) Discount() *Discount          { return p.discount }
func (p *Product) Status() ProductStatus        { return p.status }
func (p *Product) CreatedAt() time.Time         { return p.createdAt }
func (p *Product) UpdatedAt() time.Time         { return p.updatedAt }
func (p *Product) ArchivedAt() *time.Time       { return p.archivedAt }
