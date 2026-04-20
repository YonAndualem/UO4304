// Package common provides base building blocks shared across all domain aggregates.
// In DDD, every aggregate root embeds this struct to gain domain event management.
package common

// AggregateRoot is the base type for every DDD aggregate root in this system.
//
// An aggregate root is the single entry point to a cluster of domain objects.
// All state changes happen through it, and it is the unit of consistency — meaning
// the invariants it enforces are always satisfied by the time a transaction completes.
//
// Domain events are collected here during a business operation and pulled out by the
// application layer after the repository persists the aggregate, so downstream
// consumers (e.g. notification services) can react without tight coupling.
type AggregateRoot struct {
	events []DomainEvent
}

// AddEvent appends a domain event to the internal list.
// Called by aggregate methods whenever a meaningful state change occurs.
func (a *AggregateRoot) AddEvent(event DomainEvent) {
	a.events = append(a.events, event)
}

// PullEvents returns all collected domain events and clears the internal list.
// The application layer calls this after a successful repository save so that
// the events can be dispatched exactly once.
func (a *AggregateRoot) PullEvents() []DomainEvent {
	events := a.events
	a.events = nil
	return events
}
