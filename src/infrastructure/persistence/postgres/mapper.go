package postgres

import (
	"github.com/enterprise/trade-license/src/domain/tradelivense"
	"github.com/enterprise/trade-license/src/infrastructure/persistence/postgres/models"
)

func toModel(app *tradelivense.TradeLicenseApplication) *models.Application {
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

func toDomain(m *models.Application) (*tradelivense.TradeLicenseApplication, error) {
	appID, err := tradelivense.ApplicationIDFrom(m.ID)
	if err != nil {
		return nil, err
	}

	licenseType, err := tradelivense.NewLicenseType(m.LicenseType)
	if err != nil {
		return nil, err
	}

	app := &tradelivense.TradeLicenseApplication{
		ID:          appID,
		LicenseType: licenseType,
		ApplicantID: m.ApplicantID,
		Status:      tradelivense.ApplicationStatus(m.Status),
		Notes:       m.Notes,
		Documents:   make([]tradelivense.Document, 0, len(m.Documents)),
		History:     make([]tradelivense.HistoryEntry, 0, len(m.History)),
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}

	if m.Commodity != nil {
		c := tradelivense.NewCommodity(m.Commodity.Name, m.Commodity.Description, m.Commodity.Category)
		c.ID = m.Commodity.ID
		app.Commodity = &c
	}

	for _, d := range m.Documents {
		doc := tradelivense.NewDocument(d.Name, d.URL, d.ContentType)
		doc.ID = d.ID
		doc.UploadedAt = d.UploadedAt
		app.Documents = append(app.Documents, doc)
	}

	if m.Payment != nil {
		p := tradelivense.NewPayment(m.Payment.Amount, m.Payment.Currency, m.Payment.TransactionID)
		p.ID = m.Payment.ID
		p.PaidAt = m.Payment.PaidAt
		p.Status = tradelivense.PaymentStatus(m.Payment.Status)
		app.Payment = &p
	}

	for _, h := range m.History {
		app.History = append(app.History, tradelivense.HistoryEntry{
			ID:         h.ID,
			ActorID:    h.ActorID,
			Action:     h.Action,
			FromStatus: tradelivense.ApplicationStatus(h.FromStatus),
			ToStatus:   tradelivense.ApplicationStatus(h.ToStatus),
			Notes:      h.Notes,
			OccurredAt: h.OccurredAt,
		})
	}

	return app, nil
}
