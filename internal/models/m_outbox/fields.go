// Package m_outbox defines Spanner table and column name constants for the
// outbox_events table.
package m_outbox

const (
	Table = "outbox_events"

	EventID     = "event_id"
	EventType   = "event_type"
	AggregateID = "aggregate_id"
	Payload     = "payload"
	Status      = "status"
	CreatedAt   = "created_at"
	ProcessedAt = "processed_at"
)
