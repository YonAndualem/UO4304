package postgres

import (
	domain "github.com/enterprise/trade-license/src/domain/models"
	"github.com/enterprise/trade-license/src/domain/valueobjects"

	"github.com/enterprise/trade-license/src/infrastructure/persistence/postgres/models"
)

func toModel(app *domain.TradeLicenseApplication) *models.Application {
	m := &models.Application{
		ID:          app.ID.String(),
		LicenseType: app.LicenseType.String(),
		ApplicantID: app.ApplicantID,
		Status:      string(app.Status),
		Notes:       app.Notes,
		CreatedAt:   app.CreatedAt,
		UpdatedAt:   app.UpdatedAt,
	}

	if app.Commodity != nil {
		m.Commodity = &models.Commodity{
			ID:            app.Commodity.ID,
			ApplicationID: app.ID.String(),
			Name:          app.Commodity.Name,
			Description:   app.Commodity.Description,
			Category:      app.Commodity.Category,
		}
	}

	for _, d := range app.Documents {
		m.Documents = append(m.Documents, models.Document{
			ID:            d.ID,
			ApplicationID: app.ID.String(),
			Name:          d.Name,
			URL:           d.URL,
			ContentType:   d.ContentType,
			UploadedAt:    d.UploadedAt,
		})
	}

	if app.Payment != nil {
		m.Payment = &models.Payment{
			ID:            app.Payment.ID,
			ApplicationID: app.ID.String(),
			Amount:        app.Payment.Amount,
			Currency:      app.Payment.Currency,
			TransactionID: app.Payment.TransactionID,
			PaidAt:        app.Payment.PaidAt,
			Status:        string(app.Payment.Status),
		}
	}

	for _, h := range app.History {
		m.History = append(m.History, models.ApplicationHistory{
			ID:            h.ID,
			ApplicationID: app.ID.String(),
			ActorID:       h.ActorID,
			Action:        h.Action,
			FromStatus:    string(h.FromStatus),
			ToStatus:      string(h.ToStatus),
			Notes:         h.Notes,
			OccurredAt:    h.OccurredAt,
		})
	}

	return m
}

func toDomain(m *models.Application) (*domain.TradeLicenseApplication, error) {
	appID, err := valueobjects.ApplicationIDFrom(m.ID)
	if err != nil {
		return nil, err
	}

	licenseType, err := valueobjects.NewLicenseType(m.LicenseType)
	if err != nil {
		return nil, err
	}

	app := &domain.TradeLicenseApplication{
		ID:          appID,
		LicenseType: licenseType,
		ApplicantID: m.ApplicantID,
		Status:      valueobjects.ApplicationStatus(m.Status),
		Notes:       m.Notes,
		Documents:   make([]domain.Document, 0, len(m.Documents)),
		History:     make([]domain.HistoryEntry, 0, len(m.History)),
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}

	if m.Commodity != nil {
		c := domain.NewCommodity(m.Commodity.Name, m.Commodity.Description, m.Commodity.Category)
		c.ID = m.Commodity.ID
		app.Commodity = &c
	}

	for _, d := range m.Documents {
		doc := domain.NewDocument(d.Name, d.URL, d.ContentType)
		doc.ID = d.ID
		doc.UploadedAt = d.UploadedAt
		app.Documents = append(app.Documents, doc)
	}

	if m.Payment != nil {
		p := domain.NewPayment(m.Payment.Amount, m.Payment.Currency, m.Payment.TransactionID)
		p.ID = m.Payment.ID
		p.PaidAt = m.Payment.PaidAt
		p.Status = domain.PaymentStatus(m.Payment.Status)
		app.Payment = &p
	}

	for _, h := range m.History {
		app.History = append(app.History, domain.HistoryEntry{
			ID:         h.ID,
			ActorID:    h.ActorID,
			Action:     h.Action,
			FromStatus: valueobjects.ApplicationStatus(h.FromStatus),
			ToStatus:   valueobjects.ApplicationStatus(h.ToStatus),
			Notes:      h.Notes,
			OccurredAt: h.OccurredAt,
		})
	}

	return app, nil
}
