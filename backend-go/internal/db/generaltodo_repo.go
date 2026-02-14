package db

import (
	"context"
	"database/sql"
	"time"
)

// --- GeneralTodoFolderRepository ---

type GeneralTodoFolderRepository struct {
	db *sql.DB
}

func NewGeneralTodoFolderRepository(db *sql.DB) *GeneralTodoFolderRepository {
	return &GeneralTodoFolderRepository{db: db}
}

const folderColumns = `
	general_todo_folder_id, member_id, name, sort_order,
	created_date, last_modified_date
`

func scanFolder(scanner interface{ Scan(dest ...interface{}) error }) (*GeneralTodoFolder, error) {
	f := &GeneralTodoFolder{}
	err := scanner.Scan(
		&f.ID, &f.MemberID, &f.Name, &f.SortOrder,
		&f.CreatedDate, &f.LastModifiedDate,
	)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func (r *GeneralTodoFolderRepository) FindByID(ctx context.Context, id int64) (*GeneralTodoFolder, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT `+folderColumns+` FROM general_todo_folder WHERE general_todo_folder_id = ?`, id,
	)
	return scanFolder(row)
}

func (r *GeneralTodoFolderRepository) FindByMemberID(ctx context.Context, memberID int64) ([]GeneralTodoFolder, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+folderColumns+` FROM general_todo_folder WHERE member_id = ? ORDER BY sort_order ASC`, memberID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var folders []GeneralTodoFolder
	for rows.Next() {
		f, err := scanFolder(rows)
		if err != nil {
			return nil, err
		}
		folders = append(folders, *f)
	}
	return folders, rows.Err()
}

func (r *GeneralTodoFolderRepository) Create(ctx context.Context, f *GeneralTodoFolder) (int64, error) {
	now := time.Now()
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO general_todo_folder (member_id, name, sort_order, created_date, last_modified_date)
		 VALUES (?, ?, ?, ?, ?)`,
		f.MemberID, f.Name, f.SortOrder, now, now,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *GeneralTodoFolderRepository) UpdateName(ctx context.Context, id int64, name string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE general_todo_folder SET name = ?, last_modified_date = ? WHERE general_todo_folder_id = ?`,
		name, time.Now(), id,
	)
	return err
}

func (r *GeneralTodoFolderRepository) UpdateSortOrder(ctx context.Context, id int64, sortOrder int) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE general_todo_folder SET sort_order = ?, last_modified_date = ? WHERE general_todo_folder_id = ?`,
		sortOrder, time.Now(), id,
	)
	return err
}

func (r *GeneralTodoFolderRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM general_todo_folder WHERE general_todo_folder_id = ?`, id)
	return err
}

// --- GeneralTodoCategoryRepository ---

type GeneralTodoCategoryRepository struct {
	db *sql.DB
}

func NewGeneralTodoCategoryRepository(db *sql.DB) *GeneralTodoCategoryRepository {
	return &GeneralTodoCategoryRepository{db: db}
}

const categoryColumns = `
	general_todo_category_id, folder_id, member_id, name, color, sort_order, view_mode,
	created_date, last_modified_date
`

func scanCategory(scanner interface{ Scan(dest ...interface{}) error }) (*GeneralTodoCategory, error) {
	c := &GeneralTodoCategory{}
	err := scanner.Scan(
		&c.ID, &c.FolderID, &c.MemberID, &c.Name, &c.Color, &c.SortOrder, &c.ViewMode,
		&c.CreatedDate, &c.LastModifiedDate,
	)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (r *GeneralTodoCategoryRepository) FindByID(ctx context.Context, id int64) (*GeneralTodoCategory, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT `+categoryColumns+` FROM general_todo_category WHERE general_todo_category_id = ?`, id,
	)
	return scanCategory(row)
}

func (r *GeneralTodoCategoryRepository) FindByFolderID(ctx context.Context, folderID int64) ([]GeneralTodoCategory, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+categoryColumns+` FROM general_todo_category WHERE folder_id = ? ORDER BY sort_order ASC`, folderID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []GeneralTodoCategory
	for rows.Next() {
		c, err := scanCategory(rows)
		if err != nil {
			return nil, err
		}
		categories = append(categories, *c)
	}
	return categories, rows.Err()
}

func (r *GeneralTodoCategoryRepository) FindByMemberID(ctx context.Context, memberID int64) ([]GeneralTodoCategory, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+categoryColumns+` FROM general_todo_category WHERE member_id = ? ORDER BY sort_order ASC`, memberID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []GeneralTodoCategory
	for rows.Next() {
		c, err := scanCategory(rows)
		if err != nil {
			return nil, err
		}
		categories = append(categories, *c)
	}
	return categories, rows.Err()
}

func (r *GeneralTodoCategoryRepository) Create(ctx context.Context, c *GeneralTodoCategory) (int64, error) {
	now := time.Now()
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO general_todo_category (folder_id, member_id, name, color, sort_order, view_mode,
		                                    created_date, last_modified_date)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		c.FolderID, c.MemberID, c.Name, c.Color, c.SortOrder, c.ViewMode,
		now, now,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *GeneralTodoCategoryRepository) UpdateName(ctx context.Context, id int64, name string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE general_todo_category SET name = ?, last_modified_date = ? WHERE general_todo_category_id = ?`,
		name, time.Now(), id,
	)
	return err
}

func (r *GeneralTodoCategoryRepository) UpdateColor(ctx context.Context, id int64, color string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE general_todo_category SET color = ?, last_modified_date = ? WHERE general_todo_category_id = ?`,
		color, time.Now(), id,
	)
	return err
}

func (r *GeneralTodoCategoryRepository) UpdateSortOrder(ctx context.Context, id int64, sortOrder int) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE general_todo_category SET sort_order = ?, last_modified_date = ? WHERE general_todo_category_id = ?`,
		sortOrder, time.Now(), id,
	)
	return err
}

func (r *GeneralTodoCategoryRepository) UpdateViewMode(ctx context.Context, id int64, viewMode string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE general_todo_category SET view_mode = ?, last_modified_date = ? WHERE general_todo_category_id = ?`,
		viewMode, time.Now(), id,
	)
	return err
}

func (r *GeneralTodoCategoryRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM general_todo_category WHERE general_todo_category_id = ?`, id)
	return err
}

// --- GeneralTodoItemRepository ---

type GeneralTodoItemRepository struct {
	db *sql.DB
}

func NewGeneralTodoItemRepository(db *sql.DB) *GeneralTodoItemRepository {
	return &GeneralTodoItemRepository{db: db}
}

const itemColumns = `
	general_todo_item_id, folder_id, category_id, member_id, status_id,
	title, description, start_date, due_date, is_all_day,
	created_date, last_modified_date
`

func scanItem(scanner interface{ Scan(dest ...interface{}) error }) (*GeneralTodoItem, error) {
	i := &GeneralTodoItem{}
	err := scanner.Scan(
		&i.ID, &i.FolderID, &i.CategoryID, &i.MemberID, &i.StatusID,
		&i.Title, &i.Description, &i.StartDate, &i.DueDate, &i.AllDay,
		&i.CreatedDate, &i.LastModifiedDate,
	)
	if err != nil {
		return nil, err
	}
	return i, nil
}

func (r *GeneralTodoItemRepository) FindByID(ctx context.Context, id int64) (*GeneralTodoItem, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT `+itemColumns+` FROM general_todo_item WHERE general_todo_item_id = ?`, id,
	)
	return scanItem(row)
}

func (r *GeneralTodoItemRepository) FindByCategoryID(ctx context.Context, categoryID int64) ([]GeneralTodoItem, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+itemColumns+` FROM general_todo_item WHERE category_id = ? ORDER BY general_todo_item_id ASC`, categoryID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []GeneralTodoItem
	for rows.Next() {
		i, err := scanItem(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *i)
	}
	return items, rows.Err()
}

func (r *GeneralTodoItemRepository) FindByFolderID(ctx context.Context, folderID int64) ([]GeneralTodoItem, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+itemColumns+` FROM general_todo_item WHERE folder_id = ? ORDER BY general_todo_item_id ASC`, folderID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []GeneralTodoItem
	for rows.Next() {
		i, err := scanItem(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *i)
	}
	return items, rows.Err()
}

func (r *GeneralTodoItemRepository) FindByMemberID(ctx context.Context, memberID int64) ([]GeneralTodoItem, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+itemColumns+` FROM general_todo_item WHERE member_id = ? ORDER BY general_todo_item_id ASC`, memberID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []GeneralTodoItem
	for rows.Next() {
		i, err := scanItem(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *i)
	}
	return items, rows.Err()
}

func (r *GeneralTodoItemRepository) Create(ctx context.Context, item *GeneralTodoItem) (int64, error) {
	now := time.Now()
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO general_todo_item (folder_id, category_id, member_id, status_id,
		                                title, description, start_date, due_date, is_all_day,
		                                created_date, last_modified_date)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		item.FolderID, item.CategoryID, item.MemberID, item.StatusID,
		item.Title, item.Description, item.StartDate, item.DueDate, item.AllDay,
		now, now,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *GeneralTodoItemRepository) UpdateTitle(ctx context.Context, id int64, title string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE general_todo_item SET title = ?, last_modified_date = ? WHERE general_todo_item_id = ?`,
		title, time.Now(), id,
	)
	return err
}

func (r *GeneralTodoItemRepository) UpdateDescription(ctx context.Context, id int64, description *string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE general_todo_item SET description = ?, last_modified_date = ? WHERE general_todo_item_id = ?`,
		description, time.Now(), id,
	)
	return err
}

func (r *GeneralTodoItemRepository) UpdateStatus(ctx context.Context, id int64, statusID int64) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE general_todo_item SET status_id = ?, last_modified_date = ? WHERE general_todo_item_id = ?`,
		statusID, time.Now(), id,
	)
	return err
}

func (r *GeneralTodoItemRepository) MoveTo(ctx context.Context, id int64, folderID, categoryID int64) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE general_todo_item SET folder_id = ?, category_id = ?, last_modified_date = ? WHERE general_todo_item_id = ?`,
		folderID, categoryID, time.Now(), id,
	)
	return err
}

func (r *GeneralTodoItemRepository) UpdateDates(ctx context.Context, id int64, startDate, dueDate *time.Time, allDay bool) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE general_todo_item SET start_date = ?, due_date = ?, is_all_day = ?, last_modified_date = ?
		 WHERE general_todo_item_id = ?`,
		startDate, dueDate, allDay, time.Now(), id,
	)
	return err
}

func (r *GeneralTodoItemRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM general_todo_item WHERE general_todo_item_id = ?`, id)
	return err
}

// --- GeneralTodoStatusRepository ---

type GeneralTodoStatusRepository struct {
	db *sql.DB
}

func NewGeneralTodoStatusRepository(db *sql.DB) *GeneralTodoStatusRepository {
	return &GeneralTodoStatusRepository{db: db}
}

const statusColumns = `general_todo_status_id, category_id, member_id, name, sort_order`

func scanStatus(scanner interface{ Scan(dest ...interface{}) error }) (*GeneralTodoStatus, error) {
	s := &GeneralTodoStatus{}
	err := scanner.Scan(&s.ID, &s.CategoryID, &s.MemberID, &s.Name, &s.SortOrder)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (r *GeneralTodoStatusRepository) FindByID(ctx context.Context, id int64) (*GeneralTodoStatus, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT `+statusColumns+` FROM general_todo_status WHERE general_todo_status_id = ?`, id,
	)
	return scanStatus(row)
}

func (r *GeneralTodoStatusRepository) FindByCategoryID(ctx context.Context, categoryID int64) ([]GeneralTodoStatus, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+statusColumns+` FROM general_todo_status WHERE category_id = ? ORDER BY sort_order ASC`, categoryID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var statuses []GeneralTodoStatus
	for rows.Next() {
		s, err := scanStatus(rows)
		if err != nil {
			return nil, err
		}
		statuses = append(statuses, *s)
	}
	return statuses, rows.Err()
}

func (r *GeneralTodoStatusRepository) Create(ctx context.Context, s *GeneralTodoStatus) (int64, error) {
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO general_todo_status (category_id, member_id, name, sort_order)
		 VALUES (?, ?, ?, ?)`,
		s.CategoryID, s.MemberID, s.Name, s.SortOrder,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *GeneralTodoStatusRepository) UpdateName(ctx context.Context, id int64, name string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE general_todo_status SET name = ? WHERE general_todo_status_id = ?`,
		name, id,
	)
	return err
}

func (r *GeneralTodoStatusRepository) UpdateSortOrder(ctx context.Context, id int64, sortOrder int) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE general_todo_status SET sort_order = ? WHERE general_todo_status_id = ?`,
		sortOrder, id,
	)
	return err
}

func (r *GeneralTodoStatusRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM general_todo_status WHERE general_todo_status_id = ?`, id)
	return err
}

func (r *GeneralTodoStatusRepository) DeleteByCategoryID(ctx context.Context, categoryID int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM general_todo_status WHERE category_id = ?`, categoryID)
	return err
}
