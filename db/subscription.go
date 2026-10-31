// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package db

import (
	"database/sql"
	"time"
)

// Subscription is a workspace prepaid token balance and Stripe customer.
type Subscription struct {
	Id                 Id
	WorkspaceId        Id
	ProviderCustomerId *string
	AITokensBalance    int64
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// SubscriptionRepository is the interface for subscription CRUD.
type SubscriptionRepository interface {
	Create(subscription *Subscription) error
	GetById(id Id) (*Subscription, error)
	GetByWorkspaceId(workspaceId Id) (*Subscription, error)
	Update(subscription *Subscription) error
	AddTokens(workspaceId Id, tokens int64) error
	ConsumeTokens(workspaceId Id, tokens int64) error
}

type SubscriptionRepositoryPostgres struct {
	db *sql.DB
}

// SubscriptionMeta is a single row in the subscriptions_meta table.
type SubscriptionMeta struct {
	Id             Id
	SubscriptionId Id
	Key            string
	Value          string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// SubscriptionMetaRepository is the interface for subscription metadata CRUD.
type SubscriptionMetaRepository interface {
	Create(id Id, key, value string) error
	Get(id Id, key string) (*SubscriptionMeta, error)
	Update(id Id, key, value string) error
	Delete(id Id, key string) error
	ListBySubscriptionId(id Id) ([]*SubscriptionMeta, error)
	Upsert(id Id, key, value string) error
}

type SubscriptionMetaRepositoryPostgres struct {
	db *sql.DB
}

// NewSubscriptionRepository returns a subscription repository.
func NewSubscriptionRepository(db *sql.DB) SubscriptionRepository {
	return &SubscriptionRepositoryPostgres{db: db}
}

// scanSubscription maps a row into a subscription. Column order must match the select list.
func scanSubscription(item *Subscription) []any {
	return []any{
		&item.Id,
		&item.WorkspaceId,
		&item.ProviderCustomerId,
		&item.AITokensBalance,
		&item.CreatedAt,
		&item.UpdatedAt,
	}
}

// Create inserts a subscription row.
func (r *SubscriptionRepositoryPostgres) Create(subscription *Subscription) error {
	id, err := NewId()
	if err != nil {
		return err
	}

	subscription.Id = id

	err = r.db.QueryRow(
		`INSERT INTO subscriptions (
			id, workspace_id, provider_customer_id, ai_tokens_balance
		)
		VALUES ($1, $2, $3, $4)
		RETURNING created_at, updated_at`,
		subscription.Id.String(),
		subscription.WorkspaceId.String(),
		subscription.ProviderCustomerId,
		subscription.AITokensBalance,
	).Scan(&subscription.CreatedAt, &subscription.UpdatedAt)

	return err
}

// GetById returns a subscription by id. TODO: remove if not used in future.
func (r *SubscriptionRepositoryPostgres) GetById(id Id) (*Subscription, error) {
	return QueryOne(
		r.db,
		`SELECT
			id, workspace_id, provider_customer_id, ai_tokens_balance, created_at, updated_at
		FROM subscriptions
		WHERE id = $1`,
		scanSubscription,
		id.String(),
	)
}

// GetByWorkspaceId returns a subscription by workspace id.
func (r *SubscriptionRepositoryPostgres) GetByWorkspaceId(workspaceId Id) (*Subscription, error) {
	return QueryOne(
		r.db,
		`SELECT
			id, workspace_id, provider_customer_id, ai_tokens_balance, created_at, updated_at
		FROM subscriptions
		WHERE workspace_id = $1`,
		scanSubscription,
		workspaceId.String(),
	)
}

// Update updates an existing subscription row.
func (r *SubscriptionRepositoryPostgres) Update(subscription *Subscription) error {
	_, err := r.db.Exec(
		`UPDATE subscriptions
		SET
			workspace_id = $1,
			provider_customer_id = $2,
			updated_at = $3
		WHERE id = $4`,
		subscription.WorkspaceId.String(),
		subscription.ProviderCustomerId,
		time.Now().UTC(),
		subscription.Id.String(),
	)

	return err
}

// AddTokens credits AI tokens to a workspace subscription.
func (r *SubscriptionRepositoryPostgres) AddTokens(workspaceId Id, tokens int64) error {
	_, err := r.db.Exec(
		`UPDATE subscriptions
		SET ai_tokens_balance = ai_tokens_balance + $1, updated_at = $2
		WHERE workspace_id = $3`,
		tokens,
		time.Now().UTC(),
		workspaceId.String(),
	)

	return err
}

// ConsumeTokens deducts AI tokens from a workspace subscription.
func (r *SubscriptionRepositoryPostgres) ConsumeTokens(workspaceId Id, tokens int64) error {
	_, err := r.db.Exec(
		`UPDATE subscriptions
		SET ai_tokens_balance = GREATEST(ai_tokens_balance - $1, 0), updated_at = $2
		WHERE workspace_id = $3`,
		tokens,
		time.Now().UTC(),
		workspaceId.String(),
	)

	return err
}

// NewSubscriptionMetaRepository returns the repository for subscription metadata.
func NewSubscriptionMetaRepository(db *sql.DB) SubscriptionMetaRepository {
	return &SubscriptionMetaRepositoryPostgres{db: db}
}

// scanSubscriptionMeta maps a row into subscription metadata. Column order must match the select list.
func scanSubscriptionMeta(meta *SubscriptionMeta) []any {
	return []any{
		&meta.Id,
		&meta.SubscriptionId,
		&meta.Key,
		&meta.Value,
		&meta.CreatedAt,
		&meta.UpdatedAt,
	}
}

// Create inserts a subscription metadata row. TODO: remove if not used in future.
func (r *SubscriptionMetaRepositoryPostgres) Create(id Id, key, value string) error {
	metaId, err := NewId()
	if err != nil {
		return err
	}

	_, err = r.db.Exec(
		`INSERT INTO subscriptions_meta (id, subscription_id, key, value)
		VALUES ($1, $2, $3, to_jsonb($4::text))`,
		metaId.String(), id.String(), key, value,
	)

	return err
}

// Get returns subscription metadata by key. TODO: remove if not used in future.
func (r *SubscriptionMetaRepositoryPostgres) Get(id Id, key string) (*SubscriptionMeta, error) {
	return QueryOne(
		r.db,
		`SELECT id, subscription_id, key, value #>> '{}', created_at, updated_at
		FROM subscriptions_meta
		WHERE subscription_id = $1 AND key = $2`,
		scanSubscriptionMeta,
		id.String(), key,
	)
}

// Update updates an existing subscription metadata row. TODO: remove if not used in future.
func (r *SubscriptionMetaRepositoryPostgres) Update(id Id, key, value string) error {
	_, err := r.db.Exec(
		`UPDATE subscriptions_meta
		SET value = to_jsonb($1::text), updated_at = $2
		WHERE subscription_id = $3 AND key = $4`,
		value, time.Now().UTC(), id.String(), key,
	)

	return err
}

// Delete deletes a subscription metadata row. TODO: remove if not used in future.
func (r *SubscriptionMetaRepositoryPostgres) Delete(id Id, key string) error {
	_, err := r.db.Exec(
		`DELETE FROM subscriptions_meta
		WHERE subscription_id = $1 AND key = $2`,
		id.String(), key,
	)

	return err
}

// ListBySubscriptionId lists subscription metadata rows by subscription id. TODO: remove if not used in future.
func (r *SubscriptionMetaRepositoryPostgres) ListBySubscriptionId(id Id) ([]*SubscriptionMeta, error) {
	return QueryMany(
		r.db,
		`SELECT id, subscription_id, key, value #>> '{}', created_at, updated_at
		FROM subscriptions_meta
		WHERE subscription_id = $1
		ORDER BY key`,
		scanSubscriptionMeta,
		id.String(),
	)
}

// Upsert creates or updates subscription metadata. TODO: remove if not used in future.
func (r *SubscriptionMetaRepositoryPostgres) Upsert(id Id, key, value string) error {
	existing, err := r.Get(id, key)
	if err != nil {
		return err
	}
	if existing == nil {
		return r.Create(id, key, value)
	}

	return r.Update(id, key, value)
}
