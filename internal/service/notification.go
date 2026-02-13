package service

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type NotificationService struct {
	db *sql.DB
}

func NewNotificationService(db *sql.DB) *NotificationService {
	return &NotificationService{db: db}
}

type NotificationResponse struct {
	NotificationID int64     `json:"notificationId"`
	Content        string    `json:"content"`
	IsRead         bool      `json:"isRead"`
	CreatedAt      time.Time `json:"createdAt"`
}

func (s *NotificationService) GetNotifications(ctx context.Context, username string) ([]NotificationResponse, error) {
	var memberID int64
	err := s.db.QueryRowContext(ctx, "SELECT id FROM member WHERE username = ?", username).Scan(&memberID)
	if err != nil {
		return nil, fmt.Errorf("member not found: %w", err)
	}

	rows, err := s.db.QueryContext(ctx,
		`SELECT id, COALESCE(content, ''), is_read, created_date
		 FROM notification
		 WHERE receiver_id = ?
		 ORDER BY created_date DESC`, memberID)
	if err != nil {
		return nil, fmt.Errorf("querying notifications: %w", err)
	}
	defer rows.Close()

	var notifications []NotificationResponse
	for rows.Next() {
		var n NotificationResponse
		if err := rows.Scan(&n.NotificationID, &n.Content, &n.IsRead, &n.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning notification: %w", err)
		}
		notifications = append(notifications, n)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating notifications: %w", err)
	}

	if notifications == nil {
		notifications = []NotificationResponse{}
	}
	return notifications, nil
}

func (s *NotificationService) MarkAsRead(ctx context.Context, username string, notificationID int64) error {
	var memberID int64
	err := s.db.QueryRowContext(ctx, "SELECT id FROM member WHERE username = ?", username).Scan(&memberID)
	if err != nil {
		return fmt.Errorf("member not found: %w", err)
	}

	result, err := s.db.ExecContext(ctx,
		`UPDATE notification SET is_read = true, last_modified_date = ?
		 WHERE id = ? AND receiver_id = ?`,
		time.Now(), notificationID, memberID)
	if err != nil {
		return fmt.Errorf("updating notification: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("checking rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("notification not found or not owned by user")
	}
	return nil
}

func (s *NotificationService) DeleteNotification(ctx context.Context, username string, notificationID int64) error {
	var memberID int64
	err := s.db.QueryRowContext(ctx, "SELECT id FROM member WHERE username = ?", username).Scan(&memberID)
	if err != nil {
		return fmt.Errorf("member not found: %w", err)
	}

	result, err := s.db.ExecContext(ctx,
		`DELETE FROM notification WHERE id = ? AND receiver_id = ?`,
		notificationID, memberID)
	if err != nil {
		return fmt.Errorf("deleting notification: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("checking rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("notification not found or not owned by user")
	}
	return nil
}

func (s *NotificationService) DeleteAllRead(ctx context.Context, username string) error {
	var memberID int64
	err := s.db.QueryRowContext(ctx, "SELECT id FROM member WHERE username = ?", username).Scan(&memberID)
	if err != nil {
		return fmt.Errorf("member not found: %w", err)
	}

	_, err = s.db.ExecContext(ctx,
		`DELETE FROM notification WHERE receiver_id = ? AND is_read = true`,
		memberID)
	if err != nil {
		return fmt.Errorf("deleting read notifications: %w", err)
	}
	return nil
}
