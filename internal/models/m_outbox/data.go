// Package m_outbox provides the Spanner data struct and column-name constants
// for the outbox_events table.
package m_outbox

import "time"

// Data is the raw Spanner row representation of an outbox event.
type Data struct {
	EventID     string
	EventType   string
	AggregateID string
	Payload     string
	Status      string
	CreatedAt   time.Time
	ProcessedAt time.Time
}
