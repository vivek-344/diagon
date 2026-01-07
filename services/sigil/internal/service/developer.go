package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/vivek-344/diagon/sigil/internal/domain"
	"github.com/vivek-344/diagon/sigil/utils"
)

type DeveloperService struct {
	repo domain.DeveloperRepository
}

func NewDeveloperService(repo domain.DeveloperRepository) *DeveloperService {
	return &DeveloperService{repo: repo}
}

func (s *DeveloperService) Create(ctx context.Context, input domain.CreateDeveloperInput, passwordHash string) (*domain.Developer, error) {
	slog.Debug("creating new developer", "email", input.Email)

	// Validate email format
	if !utils.IsValidEmail(input.Email) {
		return nil, domain.ErrInvalidEmail
	}

	// Validate password strength
	if err := utils.IsStrongPassword(input.Password); err != nil {
		return nil, err
	}

	// Hash the password
	passwordHash, err := utils.HashPassword(input.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}
	input.Password = ""

	dev, err := s.repo.Create(ctx, &input, passwordHash)
	if err != nil {
		if err == domain.ErrEmailExists {
			return nil, domain.ErrEmailExists
		}
		return nil, fmt.Errorf("failed to create developer: %w", err)
	}

	slog.Info("new developer created", "developer_id", dev.ID)
	return dev, nil
}

func (s *DeveloperService) VerifyEmail(ctx context.Context, id uuid.UUID) error {
	slog.Debug("verifying developer email", "developer_id", id)

	err := s.repo.VerifyEmail(ctx, id)
	if err != nil {
		if err == domain.ErrNotFound {
			return err
		}
		return fmt.Errorf("verification failed: %w", err)
	}
	slog.Debug("developer email verified successful", "developer_id", id)
	return nil
}

func (s *DeveloperService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Developer, error) {
	slog.Debug("fetching developer by ID", "developer_id", id)
	dev, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if err == domain.ErrNotFound {
			return nil, err
		}
		return nil, fmt.Errorf("failed to fetch developer: %w", err)
	}
	slog.Debug("developer fetched successfully", "data", dev)
	return dev, nil
}

func (s *DeveloperService) GetByEmail(ctx context.Context, email string) (*domain.Developer, error) {
	slog.Debug("fetching developer by email", "email", email)
	dev, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		if err == domain.ErrNotFound {
			return nil, err
		}
		return nil, fmt.Errorf("failed to fetch developer: %w", err)
	}

	slog.Debug("developer fetched successfully", "data", dev)
	return dev, nil
}

func (s *DeveloperService) GetAll(ctx context.Context, filter domain.DeveloperFilter, page int, pageSize int) ([]*domain.Developer, error) {
	slog.Debug("fetching all developers", "filter", filter, "page", page, "page_size", pageSize)
	res, err := s.repo.GetAll(ctx, filter, page, pageSize)
	if err != nil {
		if err == domain.ErrNotFound {
			return nil, err
		}
		return nil, fmt.Errorf("failed to fetch developers: %w", err)
	}
	slog.Debug("developers fetched successfully", "data", res)
	return res, nil
}

func (s *DeveloperService) UpdatePassword(ctx context.Context, id uuid.UUID, oldPassword string, newPassword string) error {
	slog.Debug("updating developer password", "developer_id", id)

	dev, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if err == domain.ErrNotFound {
			return err
		}
		return fmt.Errorf("password update failed: %w", err)
	}
	if !utils.CheckPasswordHash(oldPassword, dev.PasswordHash) {
		return domain.ErrInvalidPassword
	}

	newHash, err := utils.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("password update failed: %w", err)
	}

	err = s.repo.UpdatePassword(ctx, id, dev.PasswordHash, newHash)
	if err != nil {
		if err == domain.ErrWrongPassword {
			return err
		}
		return fmt.Errorf("password update failed: %w", err)
	}

	slog.Debug("developer updated password", "developer_id", id)
	return nil
}

func (s *DeveloperService) Update(ctx context.Context, id uuid.UUID, input *domain.UpdateDeveloperInput) error {
	slog.Debug("updating developer info", "developer_id", id)
	err := s.repo.Update(ctx, id, input)
	if err != nil {
		if err == domain.ErrNotFound {
			return err
		}
		return fmt.Errorf("failed to update developer info: %w", err)
	}
	slog.Debug("developer info updated", "developer_id", id)
	return nil
}

func (s *DeveloperService) UpdateLastLogin(ctx context.Context, id uuid.UUID) error {
	slog.Debug("updating developer last login", "developer_id", id)
	err := s.repo.UpdateLastLogin(ctx, id, time.Now())
	if err != nil {
		if err == domain.ErrNotFound {
			return err
		}
		return fmt.Errorf("failed to update last login: %w", err)
	}
	slog.Info("developer logged in", "developer_id", id)
	return nil
}

func (s *DeveloperService) ResetPassword(ctx context.Context, id uuid.UUID, newPassword string) error {
	slog.Debug("resetting developer password", "developer_id", id)
	newHash, err := utils.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("password reset failed: %w", err)
	}
	err = s.repo.ResetPassword(ctx, id, newHash)
	if err != nil {
		if err == domain.ErrNotFound {
			return err
		}
		return fmt.Errorf("password reset failed: %w", err)
	}
	slog.Debug("developer password reset successful", "developer_id", id)
	return nil
}

func (s *DeveloperService) AddMetadata(ctx context.Context, id uuid.UUID, key string, value any) error {
	slog.Debug("adding metadata to developer", "developer_id", id, "key", key)
	err := s.repo.AddMetadata(ctx, id, key, value)
	if err != nil {
		if err == domain.ErrNotFound {
			return err
		}
		return fmt.Errorf("failed to add metadata: %w", err)
	}
	slog.Debug("metadata added", "developer_id", id, "key", key)
	return nil
}

func (s *DeveloperService) Delete(ctx context.Context, id uuid.UUID) error {
	slog.Debug("deleting developer", "developer_id", id)
	err := s.repo.Delete(ctx, id)
	if err != nil {
		if err == domain.ErrNotFound {
			return err
		}
		return fmt.Errorf("failed to delete developer: %w", err)
	}
	slog.Info("developer deleted", "developer_id", id)
	return nil
}

func (s *DeveloperService) SoftDelete(ctx context.Context, id uuid.UUID) error {
	slog.Debug("soft deleting developer", "developer_id", id)
	err := s.repo.SoftDelete(ctx, id)
	if err != nil {
		if err == domain.ErrNotFound {
			return err
		}
		return fmt.Errorf("failed to soft delete developer: %w", err)
	}
	slog.Info("developer toggled deleted", "developer_id", id)
	return nil
}

func (s *DeveloperService) Suspend(ctx context.Context, id uuid.UUID) error {
	slog.Debug("suspending developer", "developer_id", id)
	err := s.repo.Suspend(ctx, id)
	if err != nil {
		if err == domain.ErrNotFound {
			return err
		}
		return fmt.Errorf("failed to suspend developer: %w", err)
	}
	slog.Info("developer suspended", "developer_id", id)
	return nil
}
