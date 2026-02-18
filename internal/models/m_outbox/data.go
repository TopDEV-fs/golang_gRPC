package m_outbox

import "time"

type Data struct {
	EventID     string
	EventType   string
	AggregateID string
	Payload     string
	Status      string
	CreatedAt   time.Time
	ProcessedAt time.Time
}
