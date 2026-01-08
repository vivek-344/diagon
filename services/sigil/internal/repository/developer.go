package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/vivek-344/diagon/sigil/internal/domain"
)

type developerRepo struct {
	db *pgxpool.Pool
}

func NewDeveloperRepository(db *pgxpool.Pool) domain.DeveloperRepository {
	return &developerRepo{db: db}
}

func (r *developerRepo) Create(ctx context.Context, input *domain.CreateDeveloperInput, passwordHash string) (*domain.Developer, error) {
	query := `
		INSERT INTO developers (
			email, password_hash, full_name, company_name
		)
		VALUES ($1, $2, $3, $4)
		RETURNING
			id, email`

	var (
		dev      domain.Developer
		metadata []byte
	)

	err := r.db.QueryRow(
		ctx, query, input.Email, passwordHash, input.FullName, input.CompanyName,
	).Scan(
		&dev.ID, &dev.Email,
	)

	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return nil, domain.ErrEmailExists
		}
		return nil, err
	}

	json.Unmarshal(metadata, &dev.Metadata)
	return &dev, nil
}

func (r *developerRepo) VerifyEmail(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE developers SET email_verified = true
		WHERE id = $1 AND status != 'deleted'`

	res, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *developerRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Developer, error) {
	query := `
		SELECT id, email, password_hash, full_name, company_name,
		       status, email_verified, plan_tier, created_at, 
		       updated_at, last_login_at, metadata
		FROM developers WHERE id = $1 AND status != 'deleted'`

	dev := &domain.Developer{}
	var metadata []byte
	var lastLogin sql.NullTime

	err := r.db.QueryRow(ctx, query, id).Scan(
		&dev.ID, &dev.Email, &dev.PasswordHash, &dev.FullName, &dev.CompanyName,
		&dev.Status, &dev.EmailVerified, &dev.PlanTier, &dev.CreatedAt,
		&dev.UpdatedAt, &lastLogin, &metadata,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	if lastLogin.Valid {
		dev.LastLoginAt = &lastLogin.Time
	}
	json.Unmarshal(metadata, &dev.Metadata)

	return dev, nil
}

func (r *developerRepo) GetByEmail(ctx context.Context, email string) (*domain.Developer, error) {
	query := `
		SELECT id, email, password_hash, full_name, company_name,
		       status, email_verified, plan_tier, created_at, 
		       updated_at, last_login_at, metadata
		FROM developers WHERE email = $1 AND status != 'deleted'`

	dev := &domain.Developer{}
	var metadata []byte
	var lastLogin sql.NullTime

	err := r.db.QueryRow(ctx, query, email).Scan(
		&dev.ID, &dev.Email, &dev.PasswordHash, &dev.FullName, &dev.CompanyName,
		&dev.Status, &dev.EmailVerified, &dev.PlanTier, &dev.CreatedAt,
		&dev.UpdatedAt, &lastLogin, &metadata,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	if lastLogin.Valid {
		dev.LastLoginAt = &lastLogin.Time
	}
	json.Unmarshal(metadata, &dev.Metadata)

	return dev, nil
}

func (r *developerRepo) GetAll(ctx context.Context, filter domain.DeveloperFilter, page int, pageSize int) ([]*domain.Developer, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	var (
		args         []any
		whereClauses []string
		argPos       = 1
	)

	switch {
	case filter.Status == nil:
		whereClauses = append(whereClauses, "status != 'deleted'")

	case *filter.Status == domain.StatusDeleted:

	default:
		whereClauses = append(whereClauses, "status = $"+fmt.Sprint(argPos))
		args = append(args, *filter.Status)
		argPos++
	}

	if filter.PlanTier != nil && *filter.PlanTier != "" {
		whereClauses = append(whereClauses, "plan_tier = $"+fmt.Sprint(argPos))
		args = append(args, *filter.PlanTier)
		argPos++
	}

	limitPos := argPos
	offsetPos := argPos + 1
	args = append(args, pageSize, (page-1)*pageSize)

	query := `
		SELECT id, email, password_hash, full_name, company_name,
		       status, email_verified, plan_tier, created_at,
		       updated_at, last_login_at, metadata
		FROM developers
	`

	if len(whereClauses) > 0 {
		query += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	query += `
		ORDER BY created_at DESC
		LIMIT $` + fmt.Sprint(limitPos) + `
		OFFSET $` + fmt.Sprint(offsetPos)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var developers []*domain.Developer

	for rows.Next() {
		dev := &domain.Developer{}
		var metadata []byte
		var lastLogin sql.NullTime

		if err := rows.Scan(
			&dev.ID,
			&dev.Email,
			&dev.PasswordHash,
			&dev.FullName,
			&dev.CompanyName,
			&dev.Status,
			&dev.EmailVerified,
			&dev.PlanTier,
			&dev.CreatedAt,
			&dev.UpdatedAt,
			&lastLogin,
			&metadata,
		); err != nil {
			return nil, err
		}

		if lastLogin.Valid {
			dev.LastLoginAt = &lastLogin.Time
		}

		json.Unmarshal(metadata, &dev.Metadata)

		developers = append(developers, dev)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(developers) == 0 {
		return nil, domain.ErrNotFound
	}

	return developers, nil
}

func (r *developerRepo) UpdatePassword(ctx context.Context, id uuid.UUID, oldPasswordHash string, newPasswordHash string) error {
	query := `
		UPDATE developers SET
			password_hash = $1,
			updated_at = NOW()
		WHERE password_hash = $2 AND id = $3 AND status != 'deleted'`

	res, err := r.db.Exec(ctx, query, newPasswordHash, oldPasswordHash, id)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return domain.ErrWrongPassword
	}
	return nil
}

func (r *developerRepo) Update(ctx context.Context, id uuid.UUID, input *domain.UpdateDeveloperInput) error {
	var updatedAt time.Time

	query := `
		UPDATE developers SET
			full_name = $1,
			company_name = $2,
			status = $3,
			plan_tier = $4,
			updated_at = NOW()
		WHERE id = $5
		RETURNING updated_at`

	err := r.db.QueryRow(ctx, query,
		input.FullName, input.CompanyName, input.Status, input.PlanTier, id,
	).Scan(&updatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		return err
	}
	return nil
}

func (r *developerRepo) UpdateLastLogin(ctx context.Context, id uuid.UUID, loginTime time.Time) error {
	query := `
		UPDATE developers SET
			last_login_at = $1,
			updated_at = NOW()
		WHERE id = $2 AND status != 'deleted'`

	res, err := r.db.Exec(ctx, query, loginTime, id)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *developerRepo) ResetPassword(ctx context.Context, id uuid.UUID, newPasswordHash string) error {
	query := `
		UPDATE developers SET
			password_hash = $1,
			updated_at = NOW()
		WHERE id = $2 AND status != 'deleted'`

	res, err := r.db.Exec(ctx, query, newPasswordHash, id)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *developerRepo) AddMetadata(ctx context.Context, id uuid.UUID, key string, value any) error {
	query := `
        UPDATE developers
        SET metadata = metadata || jsonb_build_object($1, to_jsonb($2)),
            updated_at = NOW()
        WHERE id = $3 AND status != 'deleted'
    `

	res, err := r.db.Exec(ctx, query, key, value, id)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *developerRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM developers WHERE id = $1`

	res, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *developerRepo) SoftDelete(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE developers SET
			status = 'deleted',
			updated_at = NOW()
		WHERE id = $1`

	res, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *developerRepo) Suspend(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE developers SET
			status = 'suspended',
			updated_at = NOW()
		WHERE id = $1`

	res, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
