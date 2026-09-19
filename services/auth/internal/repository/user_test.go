package repository

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/vivek-344/diagon/services/auth/internal/domain"
)

func TestUserRepository_FindByEmail(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repo := NewUserRepository(db)

	createdAt := time.Now()
	updatedAt := createdAt

	mock.ExpectQuery(
		regexp.QuoteMeta(`
			SELECT
				id,
				email,
				password_hash,
				created_at,
				updated_at
			FROM auth.users
			WHERE email = $1
		`),
	).
		WithArgs("user@example.com").
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id",
				"email",
				"password_hash",
				"created_at",
				"updated_at",
			}).
				AddRow(
					"user-123",
					"user@example.com",
					"hash",
					createdAt,
					updatedAt,
				),
		)

	user, err := repo.FindByEmail(
		context.Background(),
		"user@example.com",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if user.ID != "user-123" {
		t.Fatalf(
			"expected user-123, got %s",
			user.ID,
		)
	}

	if user.Email != "user@example.com" {
		t.Fatalf(
			"unexpected email: %s",
			user.Email,
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf(
			"unmet SQL expectations: %v",
			err,
		)
	}
}

func TestUserRepository_FindByEmail_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repo := NewUserRepository(db)

	mock.ExpectQuery(
		regexp.QuoteMeta(`
			SELECT
				id,
				email,
				password_hash,
				created_at,
				updated_at
			FROM auth.users
			WHERE email = $1
		`),
	).
		WithArgs("missing@example.com").
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id",
				"email",
				"password_hash",
				"created_at",
				"updated_at",
			}),
		)

	_, err = repo.FindByEmail(
		context.Background(),
		"missing@example.com",
	)

	if !errors.Is(err, ErrUserNotFound) {
		t.Fatalf(
			"expected ErrUserNotFound, got %v",
			err,
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf(
			"unmet SQL expectations: %v",
			err,
		)
	}
}

func TestUserRepository_Create_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repo := NewUserRepository(db)

	createdAt := time.Now()

	user := &domain.User{
		Email:        "user@example.com",
		PasswordHash: "hash",
	}

	query := regexp.QuoteMeta(`
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
	`)

	// First create: user is created successfully.
	mock.ExpectQuery(query).
		WithArgs("user@example.com", "hash").
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id",
				"email",
				"password_hash",
				"created_at",
				"updated_at",
			}).AddRow(
				"user-123",
				"user@example.com",
				"hash",
				createdAt,
				createdAt,
			),
		)

	createdUser, err := repo.Create(
		context.Background(),
		user.Email,
		user.PasswordHash,
	)
	if err != nil {
		t.Fatalf("unexpected error creating user: %v", err)
	}

	if createdUser == nil {
		t.Fatal("expected created user, got nil")
	}

	if createdUser.ID != "user-123" {
		t.Fatalf("expected user-123, got %s", createdUser.ID)
	}

	if createdUser.Email != "user@example.com" {
		t.Fatalf("unexpected email: %s", createdUser.Email)
	}

	if createdUser.PasswordHash != "hash" {
		t.Fatalf("unexpected password hash: %s", createdUser.PasswordHash)
	}

	// Verify all expected SQL queries were executed.
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestUserRepository_Create_DuplicateEmail(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repo := NewUserRepository(db)

	query := regexp.QuoteMeta(`
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
	`)

	mock.ExpectQuery(query).
		WithArgs("user@example.com", "hash").
		WillReturnError(&pgconn.PgError{
			Code:           "23505",
			ConstraintName: "users_email_unique",
		})

	user, err := repo.Create(
		context.Background(),
		"user@example.com",
		"hash",
	)

	if user != nil {
		t.Fatalf("expected nil user, got %+v", user)
	}

	if !errors.Is(err, ErrUserAlreadyExists) {
		t.Fatalf(
			"expected ErrUserAlreadyExists, got %v",
			err,
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestUserRepository_Create_DatabaseError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repo := NewUserRepository(db)

	query := regexp.QuoteMeta(`
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
	`)

	mock.ExpectQuery(query).
		WithArgs("user@example.com", "hash").
		WillReturnError(errors.New("database unavailable"))

	user, err := repo.Create(
		context.Background(),
		"user@example.com",
		"hash",
	)

	if user != nil {
		t.Fatalf("expected nil user, got %+v", user)
	}

	if err == nil {
		t.Fatal("expected database error")
	}

	if !strings.Contains(err.Error(), "repository: create user") {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestUserRepository_FindByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repo := NewUserRepository(db)

	createdAt := time.Now()
	updatedAt := createdAt

	mock.ExpectQuery(
		regexp.QuoteMeta(`
			SELECT
				id,
				email,
				password_hash,
				created_at,
				updated_at
			FROM auth.users
			WHERE id = $1
		`),
	).
		WithArgs("user-123").
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id",
				"email",
				"password_hash",
				"created_at",
				"updated_at",
			}).
				AddRow(
					"user-123",
					"user@example.com",
					"hash",
					createdAt,
					updatedAt,
				),
		)

	user, err := repo.FindByID(
		context.Background(),
		"user-123",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if user.ID != "user-123" {
		t.Fatalf(
			"expected user-123, got %s",
			user.ID,
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf(
			"unmet SQL expectations: %v",
			err,
		)
	}
}
