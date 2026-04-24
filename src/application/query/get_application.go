// Package query contains the read-side use cases of the application layer.
//
// In CQRS, queries never mutate state — they only read and return data.
// Handlers here map domain aggregates to DTOs (Data Transfer Objects), which
// are flat, serialisable structs designed for the API consumer, not for the domain.
// This separation means the API contract can evolve independently of the domain model.
package query

import (
	"context"
	"time"

	"github.com/enterprise/trade-license/src/domain/models"
	"github.com/enterprise/trade-license/src/domain/repositories"
	"github.com/enterprise/trade-license/src/domain/valueobjects"
)

// ApplicationDTO is the read model (view model) returned to API consumers.
// It is a denormalised, JSON-friendly representation of the aggregate.
// Fields that may be absent (Commodity, Payment) use pointer types so they
// serialise to null rather than empty objects.
type ApplicationDTO struct {
	ID          string            `json:"id"`
	LicenseType string            `json:"license_type"`
	ApplicantID string            `json:"applicant_id"`
	Status      string            `json:"status"`
	Notes       string            `json:"notes"`
	Commodity   *CommodityDTO     `json:"commodity,omitempty"`
	Documents   []DocumentDTO     `json:"documents"`
	Payment     *PaymentDTO       `json:"payment,omitempty"`
	History     []HistoryEntryDTO `json:"history"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

type HistoryEntryDTO struct {
	ID         string    `json:"id"`
	ActorID    string    `json:"actor_id"`
	Action     string    `json:"action"`
	FromStatus string    `json:"from_status"`
	ToStatus   string    `json:"to_status"`
	Notes      string    `json:"notes"`
	OccurredAt time.Time `json:"occurred_at"`
}

type CommodityDTO struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `json:"category"`
}

type DocumentDTO struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	URL         string    `json:"url"`
	ContentType string    `json:"content_type"`
	UploadedAt  time.Time `json:"uploaded_at"`
}

type PaymentDTO struct {
	ID            string  `json:"id"`
	Amount        float64 `json:"amount"`
	Currency      string  `json:"currency"`
	TransactionID string  `json:"transaction_id"`
	Status        string  `json:"status"`
}

// ─── GetApplicationHandler ───────────────────────────────────────────────────

// GetApplicationHandler retrieves a single application by ID.
// Used by all three roles to view application details (Steps 1–3 in each workflow).
type GetApplicationHandler struct {
	repo repositories.ApplicationRepository
}

func NewGetApplicationHandler(repo repositories.ApplicationRepository) *GetApplicationHandler {
	return &GetApplicationHandler{repo: repo}
}

func (h *GetApplicationHandler) Handle(ctx context.Context, applicationID string) (*ApplicationDTO, error) {
	appID, err := valueobjects.ApplicationIDFrom(applicationID)
	if err != nil {
		return nil, err
	}

	app, err := h.repo.FindByID(ctx, appID)
	if err != nil {
		return nil, err
	}

	return toDTO(app), nil
}

// ─── ListByStatusHandler ─────────────────────────────────────────────────────

// ListByStatusHandler returns all applications in a given workflow status.
// Reviewers use this to list SUBMITTED|REREVIEW applications; approvers use it
// for ACCEPTED applications.
type ListByStatusHandler struct {
	repo repositories.ApplicationRepository
}

func NewListByStatusHandler(repo repositories.ApplicationRepository) *ListByStatusHandler {
	return &ListByStatusHandler{repo: repo}
}

func (h *ListByStatusHandler) Handle(ctx context.Context, status string) ([]*ApplicationDTO, error) {
	apps, err := h.repo.FindByStatus(ctx, valueobjects.ApplicationStatus(status))
	if err != nil {
		return nil, err
	}

	dtos := make([]*ApplicationDTO, 0, len(apps))
	for _, app := range apps {
		dtos = append(dtos, toDTO(app))
	}
	return dtos, nil
}

// ─── ListByApplicantHandler ──────────────────────────────────────────────────

// ListByApplicantHandler returns all applications submitted by a specific customer.
type ListByApplicantHandler struct {
	repo repositories.ApplicationRepository
}

func NewListByApplicantHandler(repo repositories.ApplicationRepository) *ListByApplicantHandler {
	return &ListByApplicantHandler{repo: repo}
}

func (h *ListByApplicantHandler) Handle(ctx context.Context, applicantID string) ([]*ApplicationDTO, error) {
	apps, err := h.repo.FindByApplicantID(ctx, applicantID)
	if err != nil {
		return nil, err
	}

	dtos := make([]*ApplicationDTO, 0, len(apps))
	for _, app := range apps {
		dtos = append(dtos, toDTO(app))
	}
	return dtos, nil
}

// ─── Mapping ─────────────────────────────────────────────────────────────────

// toDTO converts a domain aggregate into a flat, JSON-serialisable DTO.
// This mapping is the boundary between the domain model and the API contract.
func toDTO(app *models.TradeLicenseApplication) *ApplicationDTO {
	history := make([]HistoryEntryDTO, 0, len(app.History))
	for _, h := range app.History {
		history = append(history, HistoryEntryDTO{
			ID:         h.ID,
			ActorID:    h.ActorID,
			Action:     h.Action,
			FromStatus: string(h.FromStatus),
			ToStatus:   string(h.ToStatus),
			Notes:      h.Notes,
			OccurredAt: h.OccurredAt,
		})
	}

	dto := &ApplicationDTO{
		ID:          app.ID.String(),
		LicenseType: app.LicenseType.String(),
		ApplicantID: app.ApplicantID,
		Status:      string(app.Status),
		Notes:       app.Notes,
		Documents:   make([]DocumentDTO, 0, len(app.Documents)),
		History:     history,
		CreatedAt:   app.CreatedAt,
		UpdatedAt:   app.UpdatedAt,
	}

	if app.Commodity != nil {
		dto.Commodity = &CommodityDTO{
			ID:          app.Commodity.ID,
			Name:        app.Commodity.Name,
			Description: app.Commodity.Description,
			Category:    app.Commodity.Category,
		}
	}

	for _, d := range app.Documents {
		dto.Documents = append(dto.Documents, DocumentDTO{
			ID:          d.ID,
			Name:        d.Name,
			URL:         d.URL,
			ContentType: d.ContentType,
			UploadedAt:  d.UploadedAt,
		})
	}

	if app.Payment != nil {
		dto.Payment = &PaymentDTO{
			ID:            app.Payment.ID,
			Amount:        app.Payment.Amount,
			Currency:      app.Payment.Currency,
			TransactionID: app.Payment.TransactionID,
			Status:        string(app.Payment.Status),
		}
	}

	return dto
}
