// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package db

import (
	"database/sql"
	"time"
)

const (
	NotificationAudienceUser      = "user"
	NotificationAudienceWorkspace = "workspace"
	NotificationAudienceSystem    = "system"

	NotificationSeverityCritical = "critical"
	NotificationSeverityError    = "error"
	NotificationSeverityWarning  = "warning"
	NotificationSeverityInfo     = "info"
)

// Notification is a single row in the notifications table.
type Notification struct {
	Id           Id
	Audience     string
	WorkspaceId  *Id
	UserId       *Id
	Type         string
	Severity     string
	Title        string
	Message      *string
	ResourceType *string
	ResourceId   *Id
	Data         *string
	Checksum     string
	ReadAt       *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// NotificationRepository is the interface for notification persistence.
type NotificationRepository interface {
	Create(notification *Notification) error
	GetById(id Id) (*Notification, error)
	GetLatestMatching(notification *Notification) (*Notification, error)
	ListForUser(workspaceId, userId Id, admin bool, limit, offset int) ([]*Notification, error)
	CountUnreadForUser(workspaceId, userId Id, admin bool) (int64, error)
	MarkRead(id, workspaceId, userId Id, admin bool) (bool, error)
	MarkAllRead(workspaceId, userId Id, admin bool) error
	Dismiss(id, workspaceId, userId Id, admin bool) (bool, error)
	Delete(id Id) error
}

type NotificationRepositoryPostgres struct {
	db *sql.DB
}

// NewNotificationRepository returns a notification repository.
func NewNotificationRepository(db *sql.DB) NotificationRepository {
	return &NotificationRepositoryPostgres{db: db}
}

// scanNotification maps a row into a notification. Column order must match the select list.
func scanNotification(notification *Notification) []any {
	return []any{
		&notification.Id,
		&notification.Audience,
		&notification.WorkspaceId,
		&notification.UserId,
		&notification.Type,
		&notification.Severity,
		&notification.Title,
		&notification.Message,
		&notification.ResourceType,
		&notification.ResourceId,
		&notification.Data,
		&notification.Checksum,
		&notification.CreatedAt,
		&notification.UpdatedAt,
	}
}

// scanNotificationRead maps a row into a notification with its read receipt. Column order must match the select list.
func scanNotificationRead(notification *Notification) []any {
	return []any{
		&notification.Id,
		&notification.Audience,
		&notification.WorkspaceId,
		&notification.UserId,
		&notification.Type,
		&notification.Severity,
		&notification.Title,
		&notification.Message,
		&notification.ResourceType,
		&notification.ResourceId,
		&notification.Data,
		&notification.Checksum,
		&notification.ReadAt,
		&notification.CreatedAt,
		&notification.UpdatedAt,
	}
}

// Create inserts a notification row.
func (r *NotificationRepositoryPostgres) Create(notification *Notification) error {
	id, err := NewId()
	if err != nil {
		return err
	}

	notification.Id = id
	if notification.Severity == "" {
		notification.Severity = NotificationSeverityInfo
	}

	err = r.db.QueryRow(
		`INSERT INTO notifications (
			id, audience, workspace_id, user_id, type, severity, title, message, resource_type, resource_id, data,
			checksum
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING created_at, updated_at`,
		notification.Id.String(),
		notification.Audience,
		notification.WorkspaceId,
		notification.UserId,
		notification.Type,
		notification.Severity,
		notification.Title,
		notification.Message,
		notification.ResourceType,
		notification.ResourceId,
		notification.Data,
		notification.Checksum,
	).Scan(&notification.CreatedAt, &notification.UpdatedAt)

	return err
}

// GetById returns a notification by id.
func (r *NotificationRepositoryPostgres) GetById(id Id) (*Notification, error) {
	return QueryOne(
		r.db,
		`SELECT
			id, audience, workspace_id, user_id, type, severity, title, message, resource_type, resource_id,
			data, checksum, created_at, updated_at
		FROM notifications
		WHERE id = $1`,
		scanNotification,
		id.String(),
	)
}

// GetLatestMatching returns the newest notification with the same target, type and resource.
func (r *NotificationRepositoryPostgres) GetLatestMatching(match *Notification) (*Notification, error) {
	return QueryOne(
		r.db,
		`SELECT
			id, audience, workspace_id, user_id, type, severity, title, message, resource_type, resource_id,
			data, checksum, created_at, updated_at
		FROM notifications
		WHERE audience = $1
			AND workspace_id IS NOT DISTINCT FROM $2
			AND user_id IS NOT DISTINCT FROM $3
			AND type = $4
			AND resource_type IS NOT DISTINCT FROM $5
			AND resource_id IS NOT DISTINCT FROM $6
		ORDER BY created_at DESC
		LIMIT 1`,
		scanNotification,
		match.Audience,
		match.WorkspaceId,
		match.UserId,
		match.Type,
		match.ResourceType,
		match.ResourceId,
	)
}

// ListForUser lists a user's undismissed notifications in a workspace, newest first.
func (r *NotificationRepositoryPostgres) ListForUser(workspaceId, userId Id, admin bool, limit, offset int) ([]*Notification, error) {
	return QueryMany(
		r.db,
		`SELECT
			n.id, n.audience, n.workspace_id, n.user_id, n.type, n.severity, n.title, n.message,
			n.resource_type, n.resource_id, n.data, n.checksum, nr.read_at, n.created_at,
			n.updated_at
		FROM notifications n
		LEFT JOIN notification_reads nr ON nr.notification_id = n.id AND nr.user_id = $2
		WHERE ((n.audience = 'workspace' AND n.workspace_id = $1)
			OR (n.audience = 'user' AND n.user_id = $2 AND (n.workspace_id IS NULL OR n.workspace_id = $1))
			OR (n.audience = 'system' AND $3::boolean))
			AND nr.is_dismissed IS NOT TRUE
		ORDER BY n.created_at DESC
		LIMIT $4 OFFSET $5`,
		scanNotificationRead,
		workspaceId.String(),
		userId.String(),
		admin,
		limit,
		offset,
	)
}

// CountUnreadForUser counts a user's unread notifications in a workspace.
func (r *NotificationRepositoryPostgres) CountUnreadForUser(workspaceId, userId Id, admin bool) (int64, error) {
	var count int64
	err := r.db.QueryRow(
		`SELECT COUNT(*)
		FROM notifications n
		LEFT JOIN notification_reads nr ON nr.notification_id = n.id AND nr.user_id = $2
		WHERE ((n.audience = 'workspace' AND n.workspace_id = $1)
			OR (n.audience = 'user' AND n.user_id = $2 AND (n.workspace_id IS NULL OR n.workspace_id = $1))
			OR (n.audience = 'system' AND $3::boolean))
			AND nr.read_at IS NULL`,
		workspaceId.String(),
		userId.String(),
		admin,
	).Scan(&count)

	return count, err
}

// MarkRead marks a notification as read by a user. It returns false when the user cannot see it.
func (r *NotificationRepositoryPostgres) MarkRead(id, workspaceId, userId Id, admin bool) (bool, error) {
	result, err := r.db.Exec(
		`INSERT INTO notification_reads (notification_id, user_id, read_at)
		SELECT n.id, $3, $5
		FROM notifications n
		WHERE n.id = $1
			AND ((n.audience = 'workspace' AND n.workspace_id = $2)
				OR (n.audience = 'user' AND n.user_id = $3 AND (n.workspace_id IS NULL OR n.workspace_id = $2))
				OR (n.audience = 'system' AND $4::boolean))
		ON CONFLICT (notification_id, user_id) DO UPDATE SET read_at = notification_reads.read_at`,
		id.String(),
		workspaceId.String(),
		userId.String(),
		admin,
		time.Now().UTC(),
	)
	if err != nil {
		return false, err
	}

	affected, err := result.RowsAffected()

	return affected > 0, err
}

// MarkAllRead marks every notification a user sees in a workspace as read.
func (r *NotificationRepositoryPostgres) MarkAllRead(workspaceId, userId Id, admin bool) error {
	_, err := r.db.Exec(
		`INSERT INTO notification_reads (notification_id, user_id, read_at)
		SELECT n.id, $2, $4
		FROM notifications n
		WHERE (n.audience = 'workspace' AND n.workspace_id = $1)
			OR (n.audience = 'user' AND n.user_id = $2 AND (n.workspace_id IS NULL OR n.workspace_id = $1))
			OR (n.audience = 'system' AND $3::boolean)
		ON CONFLICT (notification_id, user_id) DO NOTHING`,
		workspaceId.String(),
		userId.String(),
		admin,
		time.Now().UTC(),
	)

	return err
}

// Dismiss hides a notification from a user and marks it read. It returns false when the user cannot see it.
func (r *NotificationRepositoryPostgres) Dismiss(id, workspaceId, userId Id, admin bool) (bool, error) {
	result, err := r.db.Exec(
		`INSERT INTO notification_reads (notification_id, user_id, read_at, is_dismissed)
		SELECT n.id, $3, $5, TRUE
		FROM notifications n
		WHERE n.id = $1
			AND ((n.audience = 'workspace' AND n.workspace_id = $2)
				OR (n.audience = 'user' AND n.user_id = $3 AND (n.workspace_id IS NULL OR n.workspace_id = $2))
				OR (n.audience = 'system' AND $4::boolean))
		ON CONFLICT (notification_id, user_id) DO UPDATE SET is_dismissed = TRUE`,
		id.String(),
		workspaceId.String(),
		userId.String(),
		admin,
		time.Now().UTC(),
	)
	if err != nil {
		return false, err
	}

	affected, err := result.RowsAffected()

	return affected > 0, err
}

// Delete deletes a notification row and its read receipts.
func (r *NotificationRepositoryPostgres) Delete(id Id) error {
	_, err := r.db.Exec(`DELETE FROM notifications WHERE id = $1`, id.String())

	return err
}
