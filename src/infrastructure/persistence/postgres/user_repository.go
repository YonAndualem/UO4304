package postgres

import (
	"errors"
	"strings"

	"gorm.io/gorm"

	"github.com/enterprise/trade-license/src/application/auth"
	"github.com/enterprise/trade-license/src/infrastructure/persistence/postgres/models"
)

// UserRepository implements auth.UserRepository using PostgreSQL via GORM.
type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(u auth.User) error {
	m := models.User{
		ID:           u.ID,
		UserID:       u.UserID,
		Role:         u.Role,
		PasswordHash: u.PasswordHash,
	}
	if err := r.db.Create(&m).Error; err != nil {
		if strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "duplicate") {
			return auth.ErrUserExists
		}
		return err
	}
	return nil
}

func (r *UserRepository) FindByUserID(userID string) (auth.User, error) {
	var m models.User
	if err := r.db.Where("user_id = ?", userID).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return auth.User{}, auth.ErrInvalidPassword
		}
		return auth.User{}, err
	}
	return auth.User{
		ID:           m.ID,
		UserID:       m.UserID,
		Role:         m.Role,
		PasswordHash: m.PasswordHash,
	}, nil
}
