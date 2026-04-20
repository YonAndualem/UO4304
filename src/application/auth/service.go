// Package auth provides user registration, login, and JWT token management.
package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserExists      = errors.New("user ID already taken")
	ErrInvalidPassword = errors.New("invalid user ID or password")
	ErrInvalidToken    = errors.New("invalid or expired token")
)

// User is the auth layer's representation of an account.
type User struct {
	ID           string
	UserID       string
	Role         string
	PasswordHash string
}

// UserRepository is the port the auth service uses to persist users.
type UserRepository interface {
	Create(u User) error
	FindByUserID(userID string) (User, error)
}

// Claims are the JWT payload fields.
type Claims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// Service handles all auth operations.
type Service struct {
	repo      UserRepository
	jwtSecret []byte
}

func NewService(repo UserRepository, jwtSecret string) *Service {
	return &Service{repo: repo, jwtSecret: []byte(jwtSecret)}
}

// Register creates a new user account. Returns ErrUserExists if the user ID is taken.
func (s *Service) Register(userID, password, role string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.repo.Create(User{
		ID:           uuid.New().String(),
		UserID:       userID,
		Role:         role,
		PasswordHash: string(hash),
	})
}

// Login verifies credentials and returns a signed JWT plus the user's role.
func (s *Service) Login(userID, password string) (token string, role string, err error) {
	u, err := s.repo.FindByUserID(userID)
	if err != nil {
		return "", "", ErrInvalidPassword
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return "", "", ErrInvalidPassword
	}
	t, err := s.sign(u)
	if err != nil {
		return "", "", err
	}
	return t, u.Role, nil
}

// ValidateToken parses and validates a JWT, returning its claims on success.
func (s *Service) ValidateToken(tokenStr string) (*Claims, error) {
	t, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return s.jwtSecret, nil
	})
	if err != nil || !t.Valid {
		return nil, ErrInvalidToken
	}
	claims, ok := t.Claims.(*Claims)
	if !ok {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

func (s *Service) sign(u User) (string, error) {
	claims := Claims{
		UserID: u.UserID,
		Role:   u.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.jwtSecret)
}
