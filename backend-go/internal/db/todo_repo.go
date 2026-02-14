package db

import (
	"context"
	"database/sql"
	"time"
)

type TodoV2Repository struct {
	db *sql.DB
}

func NewTodoV2Repository(db *sql.DB) *TodoV2Repository {
	return &TodoV2Repository{db: db}
}

const todoV2Columns = `
	todo_id, content_id, is_checked, gold, character_gold, message,
	character_id, cool_time, sort_number, gold_check, more_reward_check,
	created_date, last_modified_date
`

func scanTodoV2(scanner interface{ Scan(dest ...interface{}) error }) (*TodoV2, error) {
	t := &TodoV2{}
	err := scanner.Scan(
		&t.ID, &t.ContentID, &t.IsChecked, &t.Gold, &t.CharacterGold, &t.Message,
		&t.CharacterID, &t.CoolTime, &t.SortNumber, &t.GoldCheck, &t.MoreRewardCheck,
		&t.CreatedDate, &t.LastModifiedDate,
	)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (r *TodoV2Repository) FindByID(ctx context.Context, id int64) (*TodoV2, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT `+todoV2Columns+` FROM todov2 WHERE todo_id = ?`, id,
	)
	return scanTodoV2(row)
}

func (r *TodoV2Repository) FindByCharacterID(ctx context.Context, characterID int64) ([]TodoV2, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+todoV2Columns+` FROM todov2 WHERE character_id = ? ORDER BY sort_number ASC`, characterID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var todos []TodoV2
	for rows.Next() {
		t, err := scanTodoV2(rows)
		if err != nil {
			return nil, err
		}
		todos = append(todos, *t)
	}
	return todos, rows.Err()
}

func (r *TodoV2Repository) FindByCharacterIDAndContentID(ctx context.Context, characterID, contentID int64) (*TodoV2, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT `+todoV2Columns+` FROM todov2 WHERE character_id = ? AND content_id = ?`,
		characterID, contentID,
	)
	return scanTodoV2(row)
}

func (r *TodoV2Repository) Create(ctx context.Context, t *TodoV2) (int64, error) {
	now := time.Now()
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO todov2 (content_id, is_checked, gold, character_gold, message,
		                      character_id, cool_time, sort_number, gold_check, more_reward_check,
		                      created_date, last_modified_date)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		t.ContentID, t.IsChecked, t.Gold, t.CharacterGold, t.Message,
		t.CharacterID, t.CoolTime, t.SortNumber, t.GoldCheck, t.MoreRewardCheck,
		now, now,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *TodoV2Repository) DeleteByID(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM todov2 WHERE todo_id = ?`, id)
	return err
}

func (r *TodoV2Repository) DeleteByCharacterID(ctx context.Context, characterID int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM todov2 WHERE character_id = ?`, characterID)
	return err
}

func (r *TodoV2Repository) UpdateCheck(ctx context.Context, id int64, isChecked bool) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE todov2 SET is_checked = ?, last_modified_date = ? WHERE todo_id = ?`,
		isChecked, time.Now(), id,
	)
	return err
}

func (r *TodoV2Repository) UpdateMessage(ctx context.Context, id int64, message *string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE todov2 SET message = ?, last_modified_date = ? WHERE todo_id = ?`,
		message, time.Now(), id,
	)
	return err
}

func (r *TodoV2Repository) UpdateSort(ctx context.Context, id int64, sortNumber int) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE todov2 SET sort_number = ?, last_modified_date = ? WHERE todo_id = ?`,
		sortNumber, time.Now(), id,
	)
	return err
}

func (r *TodoV2Repository) UpdateGoldCheck(ctx context.Context, id int64, goldCheck bool) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE todov2 SET gold_check = ?, last_modified_date = ? WHERE todo_id = ?`,
		goldCheck, time.Now(), id,
	)
	return err
}

func (r *TodoV2Repository) UpdateMoreRewardCheck(ctx context.Context, id int64, moreRewardCheck bool) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE todov2 SET more_reward_check = ?, last_modified_date = ? WHERE todo_id = ?`,
		moreRewardCheck, time.Now(), id,
	)
	return err
}

func (r *TodoV2Repository) UpdateGold(ctx context.Context, id int64, gold int, characterGold int) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE todov2 SET gold = ?, character_gold = ?, last_modified_date = ? WHERE todo_id = ?`,
		gold, characterGold, time.Now(), id,
	)
	return err
}

func (r *TodoV2Repository) UpdateContentID(ctx context.Context, id int64, contentID int64, gold int, characterGold int, coolTime int) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE todov2 SET content_id = ?, gold = ?, character_gold = ?, cool_time = ?, last_modified_date = ? WHERE todo_id = ?`,
		contentID, gold, characterGold, coolTime, time.Now(), id,
	)
	return err
}

// ResetTodoV2 resets all weekly todo checks for a given member.
func (r *TodoV2Repository) ResetTodoV2(ctx context.Context, memberID int64) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE todov2 t
		 JOIN characters c ON t.character_id = c.characters_id
		 SET t.is_checked = false, t.last_modified_date = ?
		 WHERE c.member_id = ? AND t.cool_time <= 1`,
		time.Now(), memberID,
	)
	return err
}

// ResetTodoV2CoolTime2 resets bi-weekly todo checks.
func (r *TodoV2Repository) ResetTodoV2CoolTime2(ctx context.Context, memberID int64) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE todov2 t
		 JOIN characters c ON t.character_id = c.characters_id
		 SET t.is_checked = false, t.last_modified_date = ?
		 WHERE c.member_id = ? AND t.cool_time = 2`,
		time.Now(), memberID,
	)
	return err
}

// ResetAllTodoV2 resets all weekly todo checks across all members (for scheduler).
func (r *TodoV2Repository) ResetAllTodoV2(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE todov2 SET is_checked = false, last_modified_date = ? WHERE cool_time <= 1`,
		time.Now(),
	)
	return err
}

// ResetAllTodoV2CoolTime2 resets all bi-weekly checks (for scheduler).
func (r *TodoV2Repository) ResetAllTodoV2CoolTime2(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE todov2 SET is_checked = false, last_modified_date = ? WHERE cool_time = 2`,
		time.Now(),
	)
	return err
}

func (r *TodoV2Repository) FindByMemberID(ctx context.Context, memberID int64) ([]TodoV2, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT t.`+todoV2Columns+`
		 FROM todov2 t
		 JOIN characters c ON t.character_id = c.characters_id
		 WHERE c.member_id = ?
		 ORDER BY t.sort_number ASC`, memberID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var todos []TodoV2
	for rows.Next() {
		t, err := scanTodoV2(rows)
		if err != nil {
			return nil, err
		}
		todos = append(todos, *t)
	}
	return todos, rows.Err()
}
