package models

import (
	"time"

	"github.com/google/uuid"

	"github.com/enterprise/trade-license/src/domain/valueobjects"
)

// HistoryEntry records one status transition in the append-only audit trail.
type HistoryEntry struct {
	ID         string
	ActorID    string
	Action     string
	FromStatus valueobjects.ApplicationStatus
	ToStatus   valueobjects.ApplicationStatus
	Notes      string
	OccurredAt time.Time
}

func NewHistoryEntry(actorID, action string, from, to valueobjects.ApplicationStatus, notes string) HistoryEntry {
	return HistoryEntry{
		ID:         uuid.New().String(),
		ActorID:    actorID,
		Action:     action,
		FromStatus: from,
		ToStatus:   to,
		Notes:      notes,
		OccurredAt: time.Now(),
	}
}
