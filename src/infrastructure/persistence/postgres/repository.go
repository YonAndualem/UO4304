package postgres

import (
	"context"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	domainerrors "github.com/enterprise/trade-license/src/domain/errors"
	domain "github.com/enterprise/trade-license/src/domain/models"
	"github.com/enterprise/trade-license/src/domain/repositories"
	"github.com/enterprise/trade-license/src/domain/valueobjects"
	"github.com/enterprise/trade-license/src/infrastructure/persistence/postgres/models"
)

type applicationRepository struct {
	db *gorm.DB
}

func NewApplicationRepository(db *gorm.DB) repositories.ApplicationRepository {
	return &applicationRepository{db: db}
}

// Save persists a new application and its initial history entries.
func (r *applicationRepository) Save(ctx context.Context, app *domain.TradeLicenseApplication) error {
	m := toModel(app)
	return r.db.WithContext(ctx).Create(m).Error
}

// Update persists the current aggregate state inside a single transaction:
// - saves the root record and associations
// - upserts any new history entries (ON CONFLICT DO NOTHING keeps it append-only)
func (r *applicationRepository) Update(ctx context.Context, app *domain.TradeLicenseApplication) error {
	m := toModel(app)

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(m).Error; err != nil {
			return err
		}

		// Full document replace (diff is not worth the complexity)
		if err := tx.Where("application_id = ?", m.ID).Delete(&models.Document{}).Error; err != nil {
			return err
		}
		if len(m.Documents) > 0 {
			if err := tx.Create(&m.Documents).Error; err != nil {
				return err
			}
		}

		if m.Commodity != nil {
			if err := tx.Save(m.Commodity).Error; err != nil {
				return err
			}
		}

		if m.Payment != nil {
			if err := tx.Save(m.Payment).Error; err != nil {
				return err
			}
		}

		// Append-only history: upsert by primary key, skip rows that already exist
		if len(m.History) > 0 {
			if err := tx.Clauses(clause.OnConflict{DoNothing: true}).
				Create(&m.History).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// FindByID loads an application with all associations including history.
func (r *applicationRepository) FindByID(ctx context.Context, id valueobjects.ApplicationID) (*domain.TradeLicenseApplication, error) {
	var m models.Application
	err := r.db.WithContext(ctx).
		Preload("Commodity").
		Preload("Documents").
		Preload("Payment").
		Preload("History", func(db *gorm.DB) *gorm.DB {
			return db.Order("occurred_at ASC")
		}).
		First(&m, "id = ?", id.String()).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domainerrors.ErrApplicationNotFound
	}
	if err != nil {
		return nil, err
	}
	return toDomain(&m)
}

// FindByApplicantID returns all non-deleted applications for a customer,
// newest first, with full history.
func (r *applicationRepository) FindByApplicantID(ctx context.Context, applicantID string) ([]*domain.TradeLicenseApplication, error) {
	var ms []models.Application
	err := r.db.WithContext(ctx).
		Preload("Commodity").
		Preload("Documents").
		Preload("Payment").
		Preload("History", func(db *gorm.DB) *gorm.DB {
			return db.Order("occurred_at ASC")
		}).
		Where("applicant_id = ?", applicantID).
		Order("created_at DESC").
		Find(&ms).Error
	if err != nil {
		return nil, err
	}
	return toDomainSlice(ms)
}

// FindByStatus returns applications in the given status for reviewer/approver queues.
func (r *applicationRepository) FindByStatus(ctx context.Context, status valueobjects.ApplicationStatus) ([]*domain.TradeLicenseApplication, error) {
	var ms []models.Application
	err := r.db.WithContext(ctx).
		Preload("Commodity").
		Preload("Documents").
		Preload("Payment").
		Preload("History", func(db *gorm.DB) *gorm.DB {
			return db.Order("occurred_at ASC")
		}).
		Where("status = ?", string(status)).
		Order("updated_at ASC"). // Oldest pending first
		Find(&ms).Error
	if err != nil {
		return nil, err
	}
	return toDomainSlice(ms)
}

// Delete soft-deletes the application record (GORM sets deleted_at).
func (r *applicationRepository) Delete(ctx context.Context, id valueobjects.ApplicationID) error {
	return r.db.WithContext(ctx).
		Where("id = ?", id.String()).
		Delete(&models.Application{}).Error
}

func toDomainSlice(ms []models.Application) ([]*domain.TradeLicenseApplication, error) {
	apps := make([]*domain.TradeLicenseApplication, 0, len(ms))
	for i := range ms {
		app, err := toDomain(&ms[i])
		if err != nil {
			return nil, err
		}
		apps = append(apps, app)
	}
	return apps, nil
}
