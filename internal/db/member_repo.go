package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type MemberRepository struct {
	db *sql.DB
}

func NewMemberRepository(db *sql.DB) *MemberRepository {
	return &MemberRepository{db: db}
}

func (r *MemberRepository) FindByID(ctx context.Context, id int64) (*Member, error) {
	m := &Member{}
	err := r.db.QueryRowContext(ctx,
		`SELECT member_id, api_key, username, auth_provider, access_key, password,
		        main_character, role, ads_date, inspection_schedule_hour,
		        created_date, last_modified_date
		 FROM member WHERE member_id = ?`, id,
	).Scan(
		&m.ID, &m.APIKey, &m.Username, &m.AuthProvider, &m.AccessKey, &m.Password,
		&m.MainCharacter, &m.Role, &m.AdsDate, &m.InspectionScheduleHour,
		&m.CreatedDate, &m.LastModifiedDate,
	)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (r *MemberRepository) FindByUsername(ctx context.Context, username string) (*Member, error) {
	m := &Member{}
	err := r.db.QueryRowContext(ctx,
		`SELECT member_id, api_key, username, auth_provider, access_key, password,
		        main_character, role, ads_date, inspection_schedule_hour,
		        created_date, last_modified_date
		 FROM member WHERE username = ?`, username,
	).Scan(
		&m.ID, &m.APIKey, &m.Username, &m.AuthProvider, &m.AccessKey, &m.Password,
		&m.MainCharacter, &m.Role, &m.AdsDate, &m.InspectionScheduleHour,
		&m.CreatedDate, &m.LastModifiedDate,
	)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (r *MemberRepository) Create(ctx context.Context, m *Member) (int64, error) {
	now := time.Now()
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO member (api_key, username, auth_provider, access_key, password,
		                     main_character, role, ads_date, inspection_schedule_hour,
		                     created_date, last_modified_date)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		m.APIKey, m.Username, m.AuthProvider, m.AccessKey, m.Password,
		m.MainCharacter, m.Role, m.AdsDate, m.InspectionScheduleHour,
		now, now,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *MemberRepository) UpdatePassword(ctx context.Context, id int64, password string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE member SET password = ?, auth_provider = 'none', last_modified_date = ? WHERE member_id = ?`,
		password, time.Now(), id,
	)
	return err
}

func (r *MemberRepository) UpdateMainCharacter(ctx context.Context, id int64, mainCharacter string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE member SET main_character = ?, last_modified_date = ? WHERE member_id = ?`,
		mainCharacter, time.Now(), id,
	)
	return err
}

func (r *MemberRepository) UpdateApiKey(ctx context.Context, id int64, apiKey string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE member SET api_key = ?, last_modified_date = ? WHERE member_id = ?`,
		apiKey, time.Now(), id,
	)
	return err
}

func (r *MemberRepository) UpdateProvider(ctx context.Context, id int64, provider string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE member SET auth_provider = ?, last_modified_date = ? WHERE member_id = ?`,
		provider, time.Now(), id,
	)
	return err
}

func (r *MemberRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM member WHERE member_id = ?`, id)
	return err
}

func (r *MemberRepository) UpdateAdsDate(ctx context.Context, id int64, adsDate time.Time) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE member SET ads_date = ?, last_modified_date = ? WHERE member_id = ?`,
		adsDate, time.Now(), id,
	)
	return err
}

func (r *MemberRepository) FindRole(ctx context.Context, id int64) (string, error) {
	var role string
	err := r.db.QueryRowContext(ctx,
		`SELECT role FROM member WHERE member_id = ?`, id,
	).Scan(&role)
	return role, err
}

func (r *MemberRepository) UpdateByAdmin(ctx context.Context, id int64, role *string, mainCharacter *string, adsDate *time.Time) error {
	query := "UPDATE member SET last_modified_date = ?"
	args := []interface{}{time.Now()}

	if role != nil {
		query += ", role = ?"
		args = append(args, *role)
	}
	if mainCharacter != nil {
		query += ", main_character = ?"
		args = append(args, *mainCharacter)
	}
	if adsDate != nil {
		query += ", ads_date = ?"
		args = append(args, *adsDate)
	}

	query += " WHERE member_id = ?"
	args = append(args, id)

	_, err := r.db.ExecContext(ctx, query, args...)
	return err
}

func (r *MemberRepository) CountAll(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM member`).Scan(&count)
	return count, err
}

func (r *MemberRepository) FindAll(ctx context.Context, offset, limit int) ([]Member, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT member_id, api_key, username, auth_provider, access_key, password,
		        main_character, role, ads_date, inspection_schedule_hour,
		        created_date, last_modified_date
		 FROM member ORDER BY member_id DESC LIMIT ? OFFSET ?`, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []Member
	for rows.Next() {
		var m Member
		if err := rows.Scan(
			&m.ID, &m.APIKey, &m.Username, &m.AuthProvider, &m.AccessKey, &m.Password,
			&m.MainCharacter, &m.Role, &m.AdsDate, &m.InspectionScheduleHour,
			&m.CreatedDate, &m.LastModifiedDate,
		); err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	return members, rows.Err()
}

func (r *MemberRepository) UpdateInspectionScheduleHour(ctx context.Context, id int64, hour int) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE member SET inspection_schedule_hour = ?, last_modified_date = ? WHERE member_id = ?`,
		hour, time.Now(), id,
	)
	return err
}

func (r *MemberRepository) FindAllByRole(ctx context.Context, role string) ([]Member, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT member_id, api_key, username, auth_provider, access_key, password,
		        main_character, role, ads_date, inspection_schedule_hour,
		        created_date, last_modified_date
		 FROM member WHERE role = ?`, role,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []Member
	for rows.Next() {
		var m Member
		if err := rows.Scan(
			&m.ID, &m.APIKey, &m.Username, &m.AuthProvider, &m.AccessKey, &m.Password,
			&m.MainCharacter, &m.Role, &m.AdsDate, &m.InspectionScheduleHour,
			&m.CreatedDate, &m.LastModifiedDate,
		); err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	return members, rows.Err()
}

func (r *MemberRepository) Search(ctx context.Context, keyword string, offset, limit int) ([]Member, int64, error) {
	countQuery := `SELECT COUNT(*) FROM member WHERE username LIKE ? OR main_character LIKE ?`
	pattern := fmt.Sprintf("%%%s%%", keyword)

	var total int64
	if err := r.db.QueryRowContext(ctx, countQuery, pattern, pattern).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.db.QueryContext(ctx,
		`SELECT member_id, api_key, username, auth_provider, access_key, password,
		        main_character, role, ads_date, inspection_schedule_hour,
		        created_date, last_modified_date
		 FROM member WHERE username LIKE ? OR main_character LIKE ?
		 ORDER BY member_id DESC LIMIT ? OFFSET ?`,
		pattern, pattern, limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var members []Member
	for rows.Next() {
		var m Member
		if err := rows.Scan(
			&m.ID, &m.APIKey, &m.Username, &m.AuthProvider, &m.AccessKey, &m.Password,
			&m.MainCharacter, &m.Role, &m.AdsDate, &m.InspectionScheduleHour,
			&m.CreatedDate, &m.LastModifiedDate,
		); err != nil {
			return nil, 0, err
		}
		members = append(members, m)
	}
	return members, total, rows.Err()
}
