package models

import (
	"time"

	"github.com/google/uuid"
)

// PaymentStatus tracks whether the fee associated with an application has been settled.
type PaymentStatus string

const (
	PaymentSettled PaymentStatus = "SETTLED"
	PaymentPending PaymentStatus = "PENDING"
	PaymentFailed  PaymentStatus = "FAILED"
)

// Payment records the fee settlement for an application.
type Payment struct {
	ID            string
	Amount        float64
	Currency      string
	TransactionID string
	PaidAt        time.Time
	Status        PaymentStatus
}

func NewPayment(amount float64, currency, transactionID string) Payment {
	return Payment{
		ID:            uuid.New().String(),
		Amount:        amount,
		Currency:      currency,
		TransactionID: transactionID,
		PaidAt:        time.Now(),
		Status:        PaymentSettled,
	}
}
