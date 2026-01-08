package domain

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

type Status string
type contextKey string

const (
	StatusPending   Status     = "pending"
	StatusActive    Status     = "active"
	StatusSuspended Status     = "suspended"
	StatusDeleted   Status     = "deleted"
	DeveloperIDKey  contextKey = "developer_id"
	EmailKey        contextKey = "email"
)

var (
	ErrEmailExists     = errors.New("email already registered")
	ErrInvalidPassword = errors.New("invalid password")
	ErrInvalidEmail    = errors.New("invalid email format")
	ErrShortPassword   = errors.New("password should be at least 8 characters long")
	ErrWeakPassword    = errors.New("password should include a letter and a number/symbol")
	ErrNotFound        = errors.New("developer not found")
	ErrWrongPassword   = errors.New("wrong password")
	ErrInvalidInput    = errors.New("invalid input")
)

type Developer struct {
	ID            uuid.UUID
	Email         string
	PasswordHash  string
	FullName      *string
	CompanyName   *string
	Status        Status
	EmailVerified bool
	PlanTier      string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	LastLoginAt   *time.Time
	Metadata      map[string]any
}

type DeveloperFilter struct {
	Status   *Status
	PlanTier *string
}

// Repository interface for Developer entity
type DeveloperRepository interface {
	Create(ctx context.Context, input *CreateDeveloperInput, passwordHash string) (*Developer, error)
	VerifyEmail(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*Developer, error)
	GetByEmail(ctx context.Context, email string) (*Developer, error)
	GetAll(ctx context.Context, filter DeveloperFilter, page int, pageSize int) ([]*Developer, error) // with filters (ie. status and plan tier)
	UpdatePassword(ctx context.Context, id uuid.UUID, oldPasswordHash string, newPasswordHash string) error
	Update(ctx context.Context, id uuid.UUID, input *UpdateDeveloperInput) error
	UpdateLastLogin(ctx context.Context, id uuid.UUID, loginTime time.Time) error
	ResetPassword(ctx context.Context, id uuid.UUID, newPasswordHash string) error
	AddMetadata(ctx context.Context, id uuid.UUID, key string, value any) error
	Delete(ctx context.Context, id uuid.UUID) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
	Suspend(ctx context.Context, id uuid.UUID) error
}

// Input DTOs
type CreateDeveloperInput struct {
	Email       string
	Password    string
	FullName    *string
	CompanyName *string
}

type UpdateDeveloperInput struct {
	FullName    *string
	CompanyName *string
	Status      *Status
	PlanTier    *string
}
