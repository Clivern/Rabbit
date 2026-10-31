// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package db

import (
	"database/sql"
	"time"

	"github.com/samber/lo"
)

// User role and provider constants
const (
	UserRoleAdmin      = "admin"
	UserRoleRegular    = "regular"
	UserRoleOwner      = "owner"
	UserRoleReadOnly   = "readonly"
	UserRoleBot        = "bot"
	UserProviderLocal  = "local"
	UserProviderGithub = "github"
	UserProviderGoogle = "google"
	UserLanguageEN     = "en"
	UserLanguageFR     = "fr"
	UserThemeDefault   = "default"
	UserThemeBlue      = "blue"
	UserThemeSlate     = "slate"
	UserThemeEmerald   = "emerald"
	UserThemeDark      = "dark"
)

// BotUserId is the ID of the platform bot user.
const BotUserId Id = "00000000-0000-0000-0000-000000000000"

// BotUserName is the name of the platform bot user.
const BotUserName = "Sloth"

// User is the DB row for a user.
type User struct {
	Id               Id
	Name             string
	Email            string
	Password         string
	Provider         string
	ProviderUserId   *string
	Role             string
	IsActive         bool
	IsEmailVerified  bool
	EmailVerifyToken *string
	LastLoginAt      time.Time
	Language         string
	Theme            string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// UserRepository is the interface for user CRUD.
type UserRepository interface {
	Create(user *User) error
	GetById(id Id) (*User, error)
	GetBot() (*User, error)
	GetByAPIKey(apiKey string) (*User, error)
	GetByEmail(email string) (*User, error)
	GetByProvider(provider, providerUserId string) (*User, error)
	Update(user *User) error
	UpdateLastLogin(id Id) error
	Delete(id Id) error
	List(limit, offset int) ([]*User, error)
	Count() (int64, error)
}

// UserRepositoryPostgres implements UserRepository for PostgreSQL
type UserRepositoryPostgres struct {
	db *sql.DB
}

// UserMeta is extra metadata for a user.
type UserMeta struct {
	Id        Id
	Key       string
	Value     string
	UserId    Id
	CreatedAt time.Time
	UpdatedAt time.Time
}

// UserMetaRepository is the interface for user metadata CRUD.
type UserMetaRepository interface {
	Create(id Id, key, value string) error
	Get(id Id, key string) (*UserMeta, error)
	Update(id Id, key, value string) error
	Delete(id Id, key string) error
	ListByUser(id Id) ([]*UserMeta, error)
	Upsert(id Id, key, value string) error
}

// UserMetaRepositoryPostgres implements UserMetaRepository for PostgreSQL
type UserMetaRepositoryPostgres struct {
	db *sql.DB
}

// NewUserRepository returns the repository for the current driver
func NewUserRepository(db *sql.DB) UserRepository {
	return &UserRepositoryPostgres{db: db}
}

// scanUser maps a row into a user. Column order must match the select list.
func scanUser(user *User) []any {
	return []any{
		&user.Id,
		&user.Name,
		&user.Email,
		&user.Password,
		&user.Provider,
		&user.ProviderUserId,
		&user.Role,
		&user.IsActive,
		&user.IsEmailVerified,
		&user.EmailVerifyToken,
		&user.LastLoginAt,
		&user.Language,
		&user.Theme,
		&user.CreatedAt,
		&user.UpdatedAt,
	}
}

// --- Postgres ---

// Create inserts a row
func (r *UserRepositoryPostgres) Create(user *User) error {
	if lo.IsEmpty(user.Id) {
		id, err := NewId()
		if err != nil {
			return err
		}

		user.Id = id
	}

	_, err := r.db.Exec(
		`INSERT INTO users (
			id, name, email, pwd_hash, provider, provider_user_id, role, is_active,
			is_email_verified, email_verify_token, last_login_at, language, theme
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`,
		user.Id.String(),
		user.Name,
		user.Email,
		user.Password,
		user.Provider,
		user.ProviderUserId,
		user.Role,
		user.IsActive,
		user.IsEmailVerified,
		user.EmailVerifyToken,
		user.LastLoginAt,
		user.Language,
		user.Theme,
	)

	return err
}

// GetById returns a user by Id
func (r *UserRepositoryPostgres) GetById(id Id) (*User, error) {
	return QueryOne(
		r.db,
		`SELECT
			id, name, email, pwd_hash, provider, provider_user_id, role, is_active,
			is_email_verified, email_verify_token, last_login_at, language, theme, created_at, updated_at
		FROM users
		WHERE id = $1`,
		scanUser,
		id.String(),
	)
}

// GetBot returns the platform bot user. TODO: remove if not used in future.
func (r *UserRepositoryPostgres) GetBot() (*User, error) {
	return r.GetById(BotUserId)
}

// GetByAPIKey returns a user by API key (raw key lookup)
func (r *UserRepositoryPostgres) GetByAPIKey(apiKey string) (*User, error) {
	if lo.IsEmpty(apiKey) {
		return nil, nil
	}

	return QueryOne(
		r.db,
		`SELECT
			u.id, u.name, u.email, u.pwd_hash, u.provider, u.provider_user_id,
			u.role, u.is_active, u.is_email_verified, u.email_verify_token,
			u.last_login_at, u.language, u.theme, u.created_at, u.updated_at
		FROM users u
		INNER JOIN user_api_keys k ON u.id = k.user_id
		WHERE k.token = $1 AND (k.expires_at IS NULL OR k.expires_at > CURRENT_TIMESTAMP)`,
		scanUser,
		apiKey,
	)
}

// GetByEmail returns a user by email
func (r *UserRepositoryPostgres) GetByEmail(email string) (*User, error) {
	return QueryOne(
		r.db,
		`SELECT
			id, name, email, pwd_hash, provider, provider_user_id, role, is_active,
			is_email_verified, email_verify_token, last_login_at, language, theme, created_at, updated_at
		FROM users
		WHERE email = $1`,
		scanUser,
		email,
	)
}

// GetByProvider returns a user by OAuth provider identity
func (r *UserRepositoryPostgres) GetByProvider(provider, providerUserId string) (*User, error) {
	return QueryOne(
		r.db,
		`SELECT
			id, name, email, pwd_hash, provider, provider_user_id, role, is_active,
			is_email_verified, email_verify_token, last_login_at, language, theme, created_at, updated_at
		FROM users
		WHERE provider = $1 AND provider_user_id = $2`,
		scanUser,
		provider,
		providerUserId,
	)
}

// Update updates a user
func (r *UserRepositoryPostgres) Update(user *User) error {
	_, err := r.db.Exec(
		`UPDATE users
		SET
			name = $1,
			email = $2,
			pwd_hash = $3,
			provider = $4,
			provider_user_id = $5,
			role = $6,
			is_active = $7,
			is_email_verified = $8,
			email_verify_token = $9,
			last_login_at = $10,
			language = $11,
			theme = $12,
			updated_at = $13
		WHERE id = $14`,
		user.Name,
		user.Email,
		user.Password,
		user.Provider,
		user.ProviderUserId,
		user.Role,
		user.IsActive,
		user.IsEmailVerified,
		user.EmailVerifyToken,
		user.LastLoginAt,
		user.Language,
		user.Theme,
		time.Now().UTC(),
		user.Id.String(),
	)

	return err
}

// UpdateLastLogin updates the last login timestamp
func (r *UserRepositoryPostgres) UpdateLastLogin(id Id) error {
	now := time.Now().UTC()
	_, err := r.db.Exec(
		`UPDATE users
		SET
			last_login_at = $1,
			updated_at = $2
		WHERE id = $3`,
		now,
		now,
		id.String(),
	)

	return err
}

// Delete removes a user. TODO: remove if not used in future.
func (r *UserRepositoryPostgres) Delete(id Id) error {
	_, err := r.db.Exec(
		`DELETE FROM users WHERE id = $1`,
		id.String(),
	)

	return err
}

// List returns a list of users. TODO: remove if not used in future.
func (r *UserRepositoryPostgres) List(limit, offset int) ([]*User, error) {
	return QueryMany(
		r.db,
		`SELECT
			id, name, email, pwd_hash, provider, provider_user_id, role, is_active,
			is_email_verified, email_verify_token, last_login_at, language, theme, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`,
		scanUser,
		limit,
		offset,
	)
}

// Count returns the total number of users
func (r *UserRepositoryPostgres) Count() (int64, error) {
	var count int64
	err := r.db.QueryRow(
		`SELECT COUNT(*)
		FROM users`,
	).Scan(&count)

	return count, err
}

// NewUserMetaRepository returns the repository for the current driver
func NewUserMetaRepository(db *sql.DB) UserMetaRepository {
	return &UserMetaRepositoryPostgres{db: db}
}

// scanUserMeta maps a row into user metadata. Column order must match the select list.
func scanUserMeta(meta *UserMeta) []any {
	return []any{
		&meta.Id,
		&meta.Key,
		&meta.Value,
		&meta.UserId,
		&meta.CreatedAt,
		&meta.UpdatedAt,
	}
}

// Create inserts a user meta. TODO: remove if not used in future.
func (r *UserMetaRepositoryPostgres) Create(id Id, key, value string) error {
	metaId, err := NewId()
	if err != nil {
		return err
	}

	_, err = r.db.Exec(
		`INSERT INTO users_meta (id, user_id, key, value)
		VALUES ($1, $2, $3, $4)`,
		metaId.String(),
		id.String(),
		key,
		value,
	)

	return err
}

// Get returns a user meta. TODO: remove if not used in future.
func (r *UserMetaRepositoryPostgres) Get(id Id, key string) (*UserMeta, error) {
	return QueryOne(
		r.db,
		`SELECT id, key, value, user_id, created_at, updated_at
		FROM users_meta
		WHERE user_id = $1 AND key = $2`,
		scanUserMeta,
		id.String(),
		key,
	)
}

// Update updates a user meta. TODO: remove if not used in future.
func (r *UserMetaRepositoryPostgres) Update(id Id, key, value string) error {
	_, err := r.db.Exec(
		`UPDATE users_meta
		SET
			value = $1,
			updated_at = $2
		WHERE user_id = $3 AND key = $4`,
		value,
		time.Now().UTC(),
		id.String(),
		key,
	)

	return err
}

// Delete removes a user meta. TODO: remove if not used in future.
func (r *UserMetaRepositoryPostgres) Delete(id Id, key string) error {
	_, err := r.db.Exec(
		`DELETE FROM users_meta WHERE user_id = $1 AND key = $2`,
		id.String(),
		key,
	)

	return err
}

// ListByUser returns a list of user meta. TODO: remove if not used in future.
func (r *UserMetaRepositoryPostgres) ListByUser(id Id) ([]*UserMeta, error) {
	return QueryMany(
		r.db,
		`SELECT id, key, value, user_id, created_at, updated_at
		FROM users_meta
		WHERE user_id = $1
		ORDER BY key`,
		scanUserMeta,
		id.String(),
	)
}

// Upsert upserts a user meta. TODO: remove if not used in future.
func (r *UserMetaRepositoryPostgres) Upsert(id Id, key, value string) error {
	existing, err := r.Get(id, key)
	if err != nil {
		return err
	}
	if existing == nil {
		return r.Create(id, key, value)
	}

	return r.Update(id, key, value)
}
