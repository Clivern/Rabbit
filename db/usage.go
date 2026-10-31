// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package db

import (
	"database/sql"
	"errors"
	"time"
)

// Usage is a single row in the usage table.
type Usage struct {
	Id          Id
	WorkspaceId Id
	Type        string
	Quantity    int64
	Unit        *string
	PeriodStart time.Time
	PeriodEnd   time.Time
	Meta        *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// UsageRepository is the interface for usage persistence.
type UsageRepository interface {
	Create(usage *Usage) error
	GetById(id Id) (*Usage, error)
	ListByWorkspaceId(workspaceId Id, limit, offset int) ([]*Usage, error)
	CountByWorkspaceId(workspaceId Id) (int64, error)
	GetQuantityByPeriod(workspaceId Id, utype string, pstart time.Time) (int64, error)
	IncrementByPeriod(workspaceId Id, utype string, pstart, pend time.Time, quantity int64, unit string) error
}

type UsageRepositoryPostgres struct {
	db *sql.DB
}

// NewUsageRepository returns a usage repository.
func NewUsageRepository(db *sql.DB) UsageRepository {
	return &UsageRepositoryPostgres{db: db}
}

// scanUsage maps a row into a usage record. Column order must match the select list.
func scanUsage(item *Usage) []any {
	return []any{
		&item.Id,
		&item.WorkspaceId,
		&item.Type,
		&item.Quantity,
		&item.Unit,
		&item.PeriodStart,
		&item.PeriodEnd,
		&item.Meta,
		&item.CreatedAt,
		&item.UpdatedAt,
	}
}

// Create inserts an usage row. TODO: remove if not used in future.
func (r *UsageRepositoryPostgres) Create(usage *Usage) error {
	id, err := NewId()
	if err != nil {
		return err
	}

	usage.Id = id

	err = r.db.QueryRow(
		`INSERT INTO usage (
			id, workspace_id, type, quantity, unit, period_start, period_end, meta
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING created_at, updated_at`,
		usage.Id.String(),
		usage.WorkspaceId.String(),
		usage.Type,
		usage.Quantity,
		usage.Unit,
		usage.PeriodStart,
		usage.PeriodEnd,
		usage.Meta,
	).Scan(&usage.CreatedAt, &usage.UpdatedAt)

	return err
}

// GetById returns an usage by id. TODO: remove if not used in future.
func (r *UsageRepositoryPostgres) GetById(id Id) (*Usage, error) {
	return QueryOne(
		r.db,
		`SELECT
			id, workspace_id, type, quantity, unit, period_start, period_end, meta,
			created_at, updated_at
		FROM usage
		WHERE id = $1`,
		scanUsage,
		id.String(),
	)
}

// ListByWorkspaceId lists usage rows by workspace id. TODO: remove if not used in future.
func (r *UsageRepositoryPostgres) ListByWorkspaceId(workspaceId Id, limit, offset int) ([]*Usage, error) {
	return QueryMany(
		r.db,
		`SELECT
			id, workspace_id, type, quantity, unit, period_start, period_end, meta,
			created_at, updated_at
		FROM usage
		WHERE workspace_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`,
		scanUsage,
		workspaceId.String(),
		limit,
		offset,
	)
}

// CountByWorkspaceId counts usage rows by workspace id. TODO: remove if not used in future.
func (r *UsageRepositoryPostgres) CountByWorkspaceId(workspaceId Id) (int64, error) {
	var count int64
	err := r.db.QueryRow(
		`SELECT COUNT(*)
		FROM usage
		WHERE workspace_id = $1`,
		workspaceId.String(),
	).Scan(&count)

	return count, err
}

// GetQuantityByPeriod returns usage quantity for a billing period.
func (r *UsageRepositoryPostgres) GetQuantityByPeriod(workspaceId Id, utype string, pstart time.Time) (int64, error) {
	var quantity int64
	err := r.db.QueryRow(
		`SELECT COALESCE(quantity, 0)
		FROM usage
		WHERE workspace_id = $1
			AND type = $2
			AND period_start = $3`,
		workspaceId.String(),
		utype,
		pstart,
	).Scan(&quantity)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}

	return quantity, err
}

// IncrementByPeriod adds quantity to the usage row for a workspace, type, and period.
func (r *UsageRepositoryPostgres) IncrementByPeriod(workspaceId Id, utype string, pstart, pend time.Time, quantity int64, unit string) error {
	if quantity == 0 {
		return nil
	}

	id, err := NewId()
	if err != nil {
		return err
	}

	_, err = r.db.Exec(
		`INSERT INTO usage (
			id, workspace_id, type, quantity, unit, period_start, period_end
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (workspace_id, type, period_start)
		DO UPDATE SET
			quantity = usage.quantity + EXCLUDED.quantity,
			updated_at = CURRENT_TIMESTAMP AT TIME ZONE 'UTC'`,
		id.String(),
		workspaceId.String(),
		utype,
		quantity,
		unit,
		pstart,
		pend,
	)

	return err
}
