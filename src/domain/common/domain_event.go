package common

import "time"

// DomainEvent represents something meaningful that happened in the domain.
//
// Events are named in the past tense (e.g. ApplicationSubmitted) because they
// record facts — things that have already occurred — not commands or intentions.
// They carry enough data for downstream consumers to act without querying back.
type DomainEvent interface {
	// EventName returns the unique, human-readable identifier for this event type.
	// Consumers use this to route or filter events.
	EventName() string

	// OccurredAt returns the wall-clock time the event was raised inside the domain.
	OccurredAt() time.Time
}

// BaseDomainEvent provides the timestamp implementation shared by all events.
// Concrete event types embed this struct so they inherit OccurredAt for free.
type BaseDomainEvent struct {
	occurredAt time.Time
}

// NewBaseDomainEvent captures the current time at the moment a domain event is created.
func NewBaseDomainEvent() BaseDomainEvent {
	return BaseDomainEvent{occurredAt: time.Now()}
}

// OccurredAt satisfies the DomainEvent interface for all embedded event types.
func (e BaseDomainEvent) OccurredAt() time.Time {
	return e.occurredAt
}
