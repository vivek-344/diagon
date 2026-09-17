package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/vivek-344/diagon/services/auth/internal/domain"
)

var ErrUserNotFound = errors.New("user not found")

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

func (r *UserRepository) Create(
	ctx context.Context,
	email string,
	passwordHash string,
) (*domain.User, error) {
	const query = `
		INSERT INTO auth.users (
			email,
			password_hash
		)
		VALUES ($1, $2)
		RETURNING
			id,
			email,
			password_hash,
			created_at,
			updated_at
	`

	var user domain.User

	err := r.db.QueryRowContext(
		ctx,
		query,
		email,
		passwordHash,
	).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("repository: create user: %w", err)
	}

	return &user, nil
}

func (r *UserRepository) FindByEmail(
	ctx context.Context,
	email string,
) (*domain.User, error) {
	const query = `
		SELECT
			id,
			email,
			password_hash,
			created_at,
			updated_at
		FROM auth.users
		WHERE email = $1
	`

	var user domain.User

	err := r.db.QueryRowContext(
		ctx,
		query,
		email,
	).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}

		return nil, fmt.Errorf("repository: find user by email: %w", err)
	}

	return &user, nil
}
