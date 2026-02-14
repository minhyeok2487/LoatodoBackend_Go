package db

import (
	"context"
	"database/sql"
	"time"
)

type FriendsRepository struct {
	db *sql.DB
}

func NewFriendsRepository(db *sql.DB) *FriendsRepository {
	return &FriendsRepository{db: db}
}

const friendsColumns = `
	friends_id, member_id, from_member, are_we_friend, ordering,
	created_date, last_modified_date,
	show_day_todo, show_raid, show_week_todo,
	check_day_todo, check_raid, check_week_todo,
	update_gauge, update_raid, setting
`

func scanFriends(scanner interface{ Scan(dest ...interface{}) error }) (*Friends, error) {
	f := &Friends{}
	err := scanner.Scan(
		&f.ID, &f.MemberID, &f.FromMember, &f.AreWeFriend, &f.Ordering,
		&f.CreatedDate, &f.LastModifiedDate,
		&f.ShowDayTodo, &f.ShowRaid, &f.ShowWeekTodo,
		&f.CheckDayTodo, &f.CheckRaid, &f.CheckWeekTodo,
		&f.UpdateGauge, &f.UpdateRaid, &f.Setting,
	)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func (r *FriendsRepository) FindByID(ctx context.Context, id int64) (*Friends, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT `+friendsColumns+` FROM friends WHERE friends_id = ?`, id,
	)
	return scanFriends(row)
}

func (r *FriendsRepository) FindByMemberID(ctx context.Context, memberID int64) ([]Friends, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+friendsColumns+` FROM friends WHERE member_id = ? ORDER BY ordering ASC`, memberID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var friends []Friends
	for rows.Next() {
		f, err := scanFriends(rows)
		if err != nil {
			return nil, err
		}
		friends = append(friends, *f)
	}
	return friends, rows.Err()
}

func (r *FriendsRepository) FindByMemberIDAndFromMember(ctx context.Context, memberID, fromMember int64) (*Friends, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT `+friendsColumns+` FROM friends WHERE member_id = ? AND from_member = ?`,
		memberID, fromMember,
	)
	return scanFriends(row)
}

func (r *FriendsRepository) Create(ctx context.Context, f *Friends) (int64, error) {
	now := time.Now()
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO friends (member_id, from_member, are_we_friend, ordering,
		                      created_date, last_modified_date,
		                      show_day_todo, show_raid, show_week_todo,
		                      check_day_todo, check_raid, check_week_todo,
		                      update_gauge, update_raid, setting)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		f.MemberID, f.FromMember, f.AreWeFriend, f.Ordering,
		now, now,
		f.ShowDayTodo, f.ShowRaid, f.ShowWeekTodo,
		f.CheckDayTodo, f.CheckRaid, f.CheckWeekTodo,
		f.UpdateGauge, f.UpdateRaid, f.Setting,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *FriendsRepository) UpdateStatus(ctx context.Context, id int64, areWeFriend bool) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE friends SET are_we_friend = ?, last_modified_date = ? WHERE friends_id = ?`,
		areWeFriend, time.Now(), id,
	)
	return err
}

func (r *FriendsRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM friends WHERE friends_id = ?`, id)
	return err
}

func (r *FriendsRepository) DeleteByMemberIDAndFromMember(ctx context.Context, memberID, fromMember int64) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM friends WHERE member_id = ? AND from_member = ?`,
		memberID, fromMember,
	)
	return err
}

func (r *FriendsRepository) UpdateSort(ctx context.Context, id int64, ordering int) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE friends SET ordering = ?, last_modified_date = ? WHERE friends_id = ?`,
		ordering, time.Now(), id,
	)
	return err
}

func (r *FriendsRepository) UpdateSettings(ctx context.Context, id int64, field string, value bool) error {
	allowedFields := map[string]bool{
		"show_day_todo": true, "show_raid": true, "show_week_todo": true,
		"check_day_todo": true, "check_raid": true, "check_week_todo": true,
		"update_gauge": true, "update_raid": true, "setting": true,
	}
	if !allowedFields[field] {
		return sql.ErrNoRows
	}
	_, err := r.db.ExecContext(ctx,
		`UPDATE friends SET `+field+` = ?, last_modified_date = ? WHERE friends_id = ?`,
		value, time.Now(), id,
	)
	return err
}

func (r *FriendsRepository) FindByFromMember(ctx context.Context, fromMemberID int64) ([]Friends, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+friendsColumns+` FROM friends WHERE from_member = ? ORDER BY ordering ASC`, fromMemberID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var friends []Friends
	for rows.Next() {
		f, err := scanFriends(rows)
		if err != nil {
			return nil, err
		}
		friends = append(friends, *f)
	}
	return friends, rows.Err()
}
