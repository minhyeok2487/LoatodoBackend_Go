package db

import (
	"context"
	"database/sql"
	"time"
)

type NotificationRepository struct {
	db *sql.DB
}

func NewNotificationRepository(db *sql.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

const notificationColumns = `
	notification_id, content, is_read, notification_type,
	board_id, comment_id, friend_id, friend_username, friend_character_name,
	community_id, inspection_character_id,
	member_id, created_date, last_modified_date
`

func scanNotification(scanner interface{ Scan(dest ...interface{}) error }) (*Notification, error) {
	n := &Notification{}
	err := scanner.Scan(
		&n.ID, &n.Content, &n.IsRead, &n.NotificationType,
		&n.BoardID, &n.CommentID, &n.FriendID, &n.FriendUsername, &n.FriendCharacterName,
		&n.CommunityID, &n.InspectionCharacterID,
		&n.ReceiverID, &n.CreatedDate, &n.LastModifiedDate,
	)
	if err != nil {
		return nil, err
	}
	return n, nil
}

func (r *NotificationRepository) FindByID(ctx context.Context, id int64) (*Notification, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT `+notificationColumns+` FROM notification WHERE notification_id = ?`, id,
	)
	return scanNotification(row)
}

func (r *NotificationRepository) FindByReceiverID(ctx context.Context, receiverID int64) ([]Notification, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+notificationColumns+` FROM notification
		 WHERE member_id = ?
		 ORDER BY created_date DESC`, receiverID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []Notification
	for rows.Next() {
		n, err := scanNotification(rows)
		if err != nil {
			return nil, err
		}
		notifications = append(notifications, *n)
	}
	return notifications, rows.Err()
}

func (r *NotificationRepository) FindUnreadByReceiverID(ctx context.Context, receiverID int64) ([]Notification, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+notificationColumns+` FROM notification
		 WHERE member_id = ? AND is_read = false
		 ORDER BY created_date DESC`, receiverID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []Notification
	for rows.Next() {
		n, err := scanNotification(rows)
		if err != nil {
			return nil, err
		}
		notifications = append(notifications, *n)
	}
	return notifications, rows.Err()
}

func (r *NotificationRepository) Create(ctx context.Context, n *Notification) (int64, error) {
	now := time.Now()
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO notification (content, is_read, notification_type,
		                           board_id, comment_id, friend_id,
		                           friend_username, friend_character_name,
		                           community_id, inspection_character_id,
		                           member_id, created_date, last_modified_date)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		n.Content, false, n.NotificationType,
		n.BoardID, n.CommentID, n.FriendID,
		n.FriendUsername, n.FriendCharacterName,
		n.CommunityID, n.InspectionCharacterID,
		n.ReceiverID, now, now,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *NotificationRepository) MarkAsRead(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE notification SET is_read = true, last_modified_date = ? WHERE notification_id = ?`,
		time.Now(), id,
	)
	return err
}

func (r *NotificationRepository) MarkAllAsRead(ctx context.Context, receiverID int64) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE notification SET is_read = true, last_modified_date = ? WHERE member_id = ? AND is_read = false`,
		time.Now(), receiverID,
	)
	return err
}

func (r *NotificationRepository) DeleteByID(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM notification WHERE notification_id = ?`, id)
	return err
}

func (r *NotificationRepository) DeleteByReceiverID(ctx context.Context, receiverID int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM notification WHERE member_id = ?`, receiverID)
	return err
}

func (r *NotificationRepository) CountUnread(ctx context.Context, receiverID int64) (int64, error) {
	var count int64
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM notification WHERE member_id = ? AND is_read = false`, receiverID,
	).Scan(&count)
	return count, err
}
