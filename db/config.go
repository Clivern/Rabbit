// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package db

import (
	"database/sql"
	"time"
)

// Config is a key-value configuration setting in the DB.
type Config struct {
	Id        Id
	Key       string
	Value     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// ConfigRepository is the interface for config CRUD.
type ConfigRepository interface {
	Create(key, value string) error
	Get(key string) (*Config, error)
	Update(key, value string) error
	Delete(key string) error
	List() ([]*Config, error)
}

// ConfigRepositoryPostgres implements ConfigRepository for PostgreSQL.
type ConfigRepositoryPostgres struct {
	db *sql.DB
}

// NewConfigRepository returns the config repository.
func NewConfigRepository(db *sql.DB) ConfigRepository {
	return &ConfigRepositoryPostgres{db: db}
}

// scanConfig maps a row into a config. Column order must match the select list.
func scanConfig(o *Config) []any {
	return []any{
		&o.Id,
		&o.Key,
		&o.Value,
		&o.CreatedAt,
		&o.UpdatedAt,
	}
}

// Create inserts a row
func (r *ConfigRepositoryPostgres) Create(key, value string) error {
	id, err := NewId()
	if err != nil {
		return err
	}

	_, err = r.db.Exec(
		`INSERT INTO configs (id, key, value)
		VALUES ($1, $2, to_jsonb($3::text))`,
		id.String(), key, value,
	)

	return err
}

// Get returns a config by key
func (r *ConfigRepositoryPostgres) Get(key string) (*Config, error) {
	return QueryOne(
		r.db,
		`SELECT id, key, value #>> '{}', created_at, updated_at
		FROM configs
		WHERE key = $1`,
		scanConfig,
		key,
	)
}

// Update updates a config
func (r *ConfigRepositoryPostgres) Update(key, value string) error {
	_, err := r.db.Exec(
		`UPDATE configs
		SET
			value = to_jsonb($1::text),
			updated_at = $2
		WHERE key = $3`,
		value, time.Now().UTC(), key,
	)

	return err
}

// Delete removes a config. TODO: remove if not used in future.
func (r *ConfigRepositoryPostgres) Delete(key string) error {
	_, err := r.db.Exec(
		`DELETE FROM configs WHERE key = $1`,
		key,
	)

	return err
}

// List returns a list of configs. TODO: remove if not used in future.
func (r *ConfigRepositoryPostgres) List() ([]*Config, error) {
	rows, err := r.db.Query(
		`SELECT id, key, value #>> '{}', created_at, updated_at
		FROM configs
		ORDER BY key`,
	)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var options []*Config
	for rows.Next() {
		o := &Config{}
		err := rows.Scan(&o.Id, &o.Key, &o.Value, &o.CreatedAt, &o.UpdatedAt)
		if err != nil {
			return nil, err
		}

		options = append(options, o)
	}

	return options, rows.Err()
}
