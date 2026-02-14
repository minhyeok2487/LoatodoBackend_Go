package db

import (
	"context"
	"database/sql"
	"time"
)

type ScheduleRepository struct {
	db *sql.DB
}

func NewScheduleRepository(db *sql.DB) *ScheduleRepository {
	return &ScheduleRepository{db: db}
}

const scheduleColumns = `
	schedule_id, character_id, schedule_raid_category, schedule_category,
	raid_name, raid_level, day_of_week, time, memo,
	repeat_week, leader, leader_schedule_id, checked, date,
	auto_check, created_date, last_modified_date
`

func scanSchedule(scanner interface{ Scan(dest ...interface{}) error }) (*Schedule, error) {
	s := &Schedule{}
	err := scanner.Scan(
		&s.ID, &s.CharacterID, &s.ScheduleRaidCategory, &s.ScheduleCategory,
		&s.RaidName, &s.RaidLevel, &s.DayOfWeek, &s.Time, &s.Memo,
		&s.RepeatWeek, &s.Leader, &s.LeaderScheduleID, &s.Checked, &s.Date,
		&s.AutoCheck, &s.CreatedDate, &s.LastModifiedDate,
	)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (r *ScheduleRepository) FindByID(ctx context.Context, id int64) (*Schedule, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT `+scheduleColumns+` FROM schedule WHERE schedule_id = ?`, id,
	)
	return scanSchedule(row)
}

func (r *ScheduleRepository) FindByCharacterID(ctx context.Context, characterID int64) ([]Schedule, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+scheduleColumns+` FROM schedule WHERE character_id = ? ORDER BY day_of_week ASC, time ASC`,
		characterID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schedules []Schedule
	for rows.Next() {
		s, err := scanSchedule(rows)
		if err != nil {
			return nil, err
		}
		schedules = append(schedules, *s)
	}
	return schedules, rows.Err()
}

func (r *ScheduleRepository) FindByCharacterIDs(ctx context.Context, characterIDs []int64) ([]Schedule, error) {
	if len(characterIDs) == 0 {
		return nil, nil
	}

	query := `SELECT ` + scheduleColumns + ` FROM schedule WHERE character_id IN (`
	args := make([]interface{}, len(characterIDs))
	for i, id := range characterIDs {
		if i > 0 {
			query += ","
		}
		query += "?"
		args[i] = id
	}
	query += `) ORDER BY day_of_week ASC, time ASC`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schedules []Schedule
	for rows.Next() {
		s, err := scanSchedule(rows)
		if err != nil {
			return nil, err
		}
		schedules = append(schedules, *s)
	}
	return schedules, rows.Err()
}

func (r *ScheduleRepository) Create(ctx context.Context, s *Schedule) (int64, error) {
	now := time.Now()
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO schedule (character_id, schedule_raid_category, schedule_category,
		                       raid_name, raid_level, day_of_week, time, memo,
		                       repeat_week, leader, leader_schedule_id, checked, date,
		                       auto_check, created_date, last_modified_date)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		s.CharacterID, s.ScheduleRaidCategory, s.ScheduleCategory,
		s.RaidName, s.RaidLevel, s.DayOfWeek, s.Time, s.Memo,
		s.RepeatWeek, s.Leader, s.LeaderScheduleID, s.Checked, s.Date,
		s.AutoCheck, now, now,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *ScheduleRepository) Update(ctx context.Context, s *Schedule) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE schedule SET
			schedule_raid_category = ?, schedule_category = ?,
			raid_name = ?, raid_level = ?, day_of_week = ?, time = ?, memo = ?,
			repeat_week = ?, leader = ?, leader_schedule_id = ?,
			checked = ?, date = ?, auto_check = ?, last_modified_date = ?
		 WHERE schedule_id = ?`,
		s.ScheduleRaidCategory, s.ScheduleCategory,
		s.RaidName, s.RaidLevel, s.DayOfWeek, s.Time, s.Memo,
		s.RepeatWeek, s.Leader, s.LeaderScheduleID,
		s.Checked, s.Date, s.AutoCheck, time.Now(), s.ID,
	)
	return err
}

func (r *ScheduleRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM schedule WHERE schedule_id = ?`, id)
	return err
}

func (r *ScheduleRepository) DeleteByLeaderScheduleID(ctx context.Context, leaderScheduleID int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM schedule WHERE leader_schedule_id = ?`, leaderScheduleID)
	return err
}

func (r *ScheduleRepository) UpdateChecked(ctx context.Context, id int64, checked bool) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE schedule SET checked = ?, last_modified_date = ? WHERE schedule_id = ?`,
		checked, time.Now(), id,
	)
	return err
}

func (r *ScheduleRepository) FindByLeaderScheduleID(ctx context.Context, leaderScheduleID int64) ([]Schedule, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+scheduleColumns+` FROM schedule WHERE leader_schedule_id = ?`, leaderScheduleID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schedules []Schedule
	for rows.Next() {
		s, err := scanSchedule(rows)
		if err != nil {
			return nil, err
		}
		schedules = append(schedules, *s)
	}
	return schedules, rows.Err()
}

func (r *ScheduleRepository) ResetCheckedByRepeatWeek(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE schedule SET checked = false, last_modified_date = ? WHERE repeat_week = true`,
		time.Now(),
	)
	return err
}

func (r *ScheduleRepository) FindAutoCheckSchedules(ctx context.Context, dayOfWeek int) ([]Schedule, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+scheduleColumns+` FROM schedule
		 WHERE auto_check = true AND repeat_week = true AND day_of_week = ? AND checked = false`,
		dayOfWeek,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schedules []Schedule
	for rows.Next() {
		s, err := scanSchedule(rows)
		if err != nil {
			return nil, err
		}
		schedules = append(schedules, *s)
	}
	return schedules, rows.Err()
}

func (r *ScheduleRepository) DeleteByCharacterID(ctx context.Context, characterID int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM schedule WHERE character_id = ?`, characterID)
	return err
}
