package service

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type GeneralTodoService struct {
	db *sql.DB
}

func NewGeneralTodoService(db *sql.DB) *GeneralTodoService {
	return &GeneralTodoService{db: db}
}

// --- Overview ---

type GeneralTodoOverview struct {
	Folders    []GeneralTodoFolderResponse    `json:"folders"`
	Categories []GeneralTodoCategoryResponse  `json:"categories"`
	Todos      []GeneralTodoItemResponse      `json:"todos"`
	Statuses   []GeneralTodoStatusResponse    `json:"statuses"`
}

type GeneralTodoFolderResponse struct {
	FolderID   int64                          `json:"folderId"`
	Name       string                         `json:"name"`
	SortNumber int                            `json:"sortNumber"`
	Categories []GeneralTodoCategoryResponse  `json:"categories"`
}

type GeneralTodoCategoryResponse struct {
	CategoryID int64                       `json:"categoryId"`
	Name       string                      `json:"name"`
	SortNumber int                         `json:"sortNumber"`
	Items      []GeneralTodoItemResponse   `json:"items"`
}

type GeneralTodoItemResponse struct {
	ItemID     int64                         `json:"itemId"`
	Name       string                        `json:"name"`
	Memo       string                        `json:"memo"`
	SortNumber int                           `json:"sortNumber"`
	Statuses   []GeneralTodoStatusResponse   `json:"statuses"`
}

type GeneralTodoStatusResponse struct {
	StatusID      int64  `json:"statusId"`
	CharacterName string `json:"characterName"`
	Checked       bool   `json:"checked"`
}

// --- Folder ---

type CreateFolderRequest struct {
	Name string `json:"name"`
}

type UpdateFolderRequest struct {
	Name string `json:"name"`
}

type FolderSortItem struct {
	FolderID   int64 `json:"folderId"`
	SortNumber int   `json:"sortNumber"`
}

// --- Category ---

type CreateCategoryRequest struct {
	FolderID int64  `json:"folderId"`
	Name     string `json:"name"`
}

type UpdateCategoryRequest struct {
	Name string `json:"name"`
}

type CategorySortItem struct {
	CategoryID int64 `json:"categoryId"`
	SortNumber int   `json:"sortNumber"`
}

// --- Item ---

type CreateItemRequest struct {
	CategoryID int64  `json:"categoryId"`
	Name       string `json:"name"`
}

type UpdateItemRequest struct {
	Name string `json:"name"`
	Memo string `json:"memo"`
}

type ItemSortItem struct {
	ItemID     int64 `json:"itemId"`
	SortNumber int   `json:"sortNumber"`
}

// --- Status ---

type CreateStatusRequest struct {
	ItemID        int64  `json:"itemId"`
	CharacterName string `json:"characterName"`
}

// getMemberID resolves the member_id for a given username.
func (s *GeneralTodoService) getMemberID(ctx context.Context, username string) (int64, error) {
	var memberID int64
	err := s.db.QueryRowContext(ctx,
		`SELECT member_id FROM member WHERE username = ?`, username,
	).Scan(&memberID)
	if err != nil {
		return 0, fmt.Errorf("member not found: %w", err)
	}
	return memberID, nil
}

// --- Overview ---

func (s *GeneralTodoService) GetOverview(ctx context.Context, username string) (*GeneralTodoOverview, error) {
	memberID, err := s.getMemberID(ctx, username)
	if err != nil {
		return nil, err
	}

	// 1. Load all folders for this member
	folderRows, err := s.db.QueryContext(ctx,
		`SELECT general_todo_folder_id, name, sort_order
		 FROM general_todo_folder WHERE member_id = ? ORDER BY sort_order ASC`, memberID,
	)
	if err != nil {
		return nil, fmt.Errorf("query folders: %w", err)
	}
	defer folderRows.Close()

	type folderRow struct {
		id        int64
		name      string
		sortOrder int
	}
	var folders []folderRow
	for folderRows.Next() {
		var f folderRow
		if err := folderRows.Scan(&f.id, &f.name, &f.sortOrder); err != nil {
			return nil, err
		}
		folders = append(folders, f)
	}
	if err := folderRows.Err(); err != nil {
		return nil, err
	}

	// 2. Load all categories for this member
	catRows, err := s.db.QueryContext(ctx,
		`SELECT general_todo_category_id, folder_id, name, sort_order
		 FROM general_todo_category WHERE member_id = ? ORDER BY sort_order ASC`, memberID,
	)
	if err != nil {
		return nil, fmt.Errorf("query categories: %w", err)
	}
	defer catRows.Close()

	type catRow struct {
		id        int64
		folderID  int64
		name      string
		sortOrder int
	}
	var categories []catRow
	for catRows.Next() {
		var c catRow
		if err := catRows.Scan(&c.id, &c.folderID, &c.name, &c.sortOrder); err != nil {
			return nil, err
		}
		categories = append(categories, c)
	}
	if err := catRows.Err(); err != nil {
		return nil, err
	}

	// 3. Load all items for this member
	itemRows, err := s.db.QueryContext(ctx,
		`SELECT general_todo_item_id, category_id, title, COALESCE(description, ''), status_id
		 FROM general_todo_item WHERE member_id = ? ORDER BY general_todo_item_id ASC`, memberID,
	)
	if err != nil {
		return nil, fmt.Errorf("query items: %w", err)
	}
	defer itemRows.Close()

	type itemRow struct {
		id         int64
		categoryID int64
		title      string
		desc       string
		statusID   int64
	}
	var items []itemRow
	for itemRows.Next() {
		var i itemRow
		if err := itemRows.Scan(&i.id, &i.categoryID, &i.title, &i.desc, &i.statusID); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := itemRows.Err(); err != nil {
		return nil, err
	}

	// 4. Load all statuses for this member
	statusRows, err := s.db.QueryContext(ctx,
		`SELECT general_todo_status_id, category_id, name, sort_order
		 FROM general_todo_status WHERE member_id = ? ORDER BY sort_order ASC`, memberID,
	)
	if err != nil {
		return nil, fmt.Errorf("query statuses: %w", err)
	}
	defer statusRows.Close()

	type statusRow struct {
		id         int64
		categoryID int64
		name       string
		sortOrder  int
	}
	var statuses []statusRow
	for statusRows.Next() {
		var st statusRow
		if err := statusRows.Scan(&st.id, &st.categoryID, &st.name, &st.sortOrder); err != nil {
			return nil, err
		}
		statuses = append(statuses, st)
	}
	if err := statusRows.Err(); err != nil {
		return nil, err
	}

	// 5. Assemble the hierarchy in memory
	// Map statuses by categoryID
	statusByCat := make(map[int64][]GeneralTodoStatusResponse)
	for _, st := range statuses {
		statusByCat[st.categoryID] = append(statusByCat[st.categoryID], GeneralTodoStatusResponse{
			StatusID:      st.id,
			CharacterName: st.name,
			Checked:       st.sortOrder != 0,
		})
	}

	// Map items by categoryID
	itemByCat := make(map[int64][]GeneralTodoItemResponse)
	for _, i := range items {
		stList := statusByCat[i.categoryID]
		if stList == nil {
			stList = []GeneralTodoStatusResponse{}
		}
		itemByCat[i.categoryID] = append(itemByCat[i.categoryID], GeneralTodoItemResponse{
			ItemID:     i.id,
			Name:       i.title,
			Memo:       i.desc,
			SortNumber: int(i.statusID),
			Statuses:   stList,
		})
	}

	// Map categories by folderID
	catByFolder := make(map[int64][]GeneralTodoCategoryResponse)
	for _, c := range categories {
		iList := itemByCat[c.id]
		if iList == nil {
			iList = []GeneralTodoItemResponse{}
		}
		catByFolder[c.folderID] = append(catByFolder[c.folderID], GeneralTodoCategoryResponse{
			CategoryID: c.id,
			Name:       c.name,
			SortNumber: c.sortOrder,
			Items:      iList,
		})
	}

	// Build folder list
	folderResponses := make([]GeneralTodoFolderResponse, 0, len(folders))
	for _, f := range folders {
		cList := catByFolder[f.id]
		if cList == nil {
			cList = []GeneralTodoCategoryResponse{}
		}
		folderResponses = append(folderResponses, GeneralTodoFolderResponse{
			FolderID:   f.id,
			Name:       f.name,
			SortNumber: f.sortOrder,
			Categories: cList,
		})
	}

	// Flatten categories, items, and statuses for top-level response
	allCategories := make([]GeneralTodoCategoryResponse, 0)
	allTodos := make([]GeneralTodoItemResponse, 0)
	allStatuses := make([]GeneralTodoStatusResponse, 0)

	for _, folder := range folderResponses {
		for _, cat := range folder.Categories {
			allCategories = append(allCategories, cat)
			for _, item := range cat.Items {
				allTodos = append(allTodos, item)
				allStatuses = append(allStatuses, item.Statuses...)
			}
		}
	}

	return &GeneralTodoOverview{
		Folders:    folderResponses,
		Categories: allCategories,
		Todos:      allTodos,
		Statuses:   allStatuses,
	}, nil
}

// --- Folder ---

func (s *GeneralTodoService) CreateFolder(ctx context.Context, username string, req CreateFolderRequest) (*GeneralTodoFolderResponse, error) {
	memberID, err := s.getMemberID(ctx, username)
	if err != nil {
		return nil, err
	}

	// Get the next sort_order
	var maxSort sql.NullInt64
	err = s.db.QueryRowContext(ctx,
		`SELECT MAX(sort_order) FROM general_todo_folder WHERE member_id = ?`, memberID,
	).Scan(&maxSort)
	if err != nil {
		return nil, fmt.Errorf("query max sort_order: %w", err)
	}
	nextSort := 0
	if maxSort.Valid {
		nextSort = int(maxSort.Int64) + 1
	}

	now := time.Now()
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO general_todo_folder (member_id, name, sort_order, created_date, last_modified_date)
		 VALUES (?, ?, ?, ?, ?)`,
		memberID, req.Name, nextSort, now, now,
	)
	if err != nil {
		return nil, fmt.Errorf("insert folder: %w", err)
	}
	folderID, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("get last insert id: %w", err)
	}

	return &GeneralTodoFolderResponse{
		FolderID:   folderID,
		Name:       req.Name,
		SortNumber: nextSort,
		Categories: []GeneralTodoCategoryResponse{},
	}, nil
}

func (s *GeneralTodoService) UpdateFolder(ctx context.Context, username string, folderID int64, req UpdateFolderRequest) error {
	memberID, err := s.getMemberID(ctx, username)
	if err != nil {
		return err
	}

	result, err := s.db.ExecContext(ctx,
		`UPDATE general_todo_folder SET name = ?, last_modified_date = ?
		 WHERE general_todo_folder_id = ? AND member_id = ?`,
		req.Name, time.Now(), folderID, memberID,
	)
	if err != nil {
		return fmt.Errorf("update folder: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("folder not found or not owned by user")
	}
	return nil
}

func (s *GeneralTodoService) DeleteFolder(ctx context.Context, username string, folderID int64) error {
	memberID, err := s.getMemberID(ctx, username)
	if err != nil {
		return err
	}

	// Delete associated statuses, items, and categories within a transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// Delete statuses for categories in this folder
	_, err = tx.ExecContext(ctx,
		`DELETE FROM general_todo_status WHERE member_id = ? AND category_id IN
		 (SELECT general_todo_category_id FROM general_todo_category WHERE folder_id = ? AND member_id = ?)`,
		memberID, folderID, memberID,
	)
	if err != nil {
		return fmt.Errorf("delete statuses: %w", err)
	}

	// Delete items in this folder
	_, err = tx.ExecContext(ctx,
		`DELETE FROM general_todo_item WHERE folder_id = ? AND member_id = ?`,
		folderID, memberID,
	)
	if err != nil {
		return fmt.Errorf("delete items: %w", err)
	}

	// Delete categories in this folder
	_, err = tx.ExecContext(ctx,
		`DELETE FROM general_todo_category WHERE folder_id = ? AND member_id = ?`,
		folderID, memberID,
	)
	if err != nil {
		return fmt.Errorf("delete categories: %w", err)
	}

	// Delete the folder itself
	result, err := tx.ExecContext(ctx,
		`DELETE FROM general_todo_folder WHERE general_todo_folder_id = ? AND member_id = ?`,
		folderID, memberID,
	)
	if err != nil {
		return fmt.Errorf("delete folder: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("folder not found or not owned by user")
	}

	return tx.Commit()
}

func (s *GeneralTodoService) UpdateFolderSort(ctx context.Context, username string, items []FolderSortItem) error {
	memberID, err := s.getMemberID(ctx, username)
	if err != nil {
		return err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	now := time.Now()
	for _, item := range items {
		_, err := tx.ExecContext(ctx,
			`UPDATE general_todo_folder SET sort_order = ?, last_modified_date = ?
			 WHERE general_todo_folder_id = ? AND member_id = ?`,
			item.SortNumber, now, item.FolderID, memberID,
		)
		if err != nil {
			return fmt.Errorf("update folder sort: %w", err)
		}
	}

	return tx.Commit()
}

// --- Category ---

func (s *GeneralTodoService) CreateCategory(ctx context.Context, username string, req CreateCategoryRequest) (*GeneralTodoCategoryResponse, error) {
	memberID, err := s.getMemberID(ctx, username)
	if err != nil {
		return nil, err
	}

	// Get the next sort_order for this folder
	var maxSort sql.NullInt64
	err = s.db.QueryRowContext(ctx,
		`SELECT MAX(sort_order) FROM general_todo_category WHERE folder_id = ? AND member_id = ?`,
		req.FolderID, memberID,
	).Scan(&maxSort)
	if err != nil {
		return nil, fmt.Errorf("query max sort_order: %w", err)
	}
	nextSort := 0
	if maxSort.Valid {
		nextSort = int(maxSort.Int64) + 1
	}

	now := time.Now()
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO general_todo_category (folder_id, member_id, name, sort_order, view_mode, created_date, last_modified_date)
		 VALUES (?, ?, ?, ?, 'LIST', ?, ?)`,
		req.FolderID, memberID, req.Name, nextSort, now, now,
	)
	if err != nil {
		return nil, fmt.Errorf("insert category: %w", err)
	}
	categoryID, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("get last insert id: %w", err)
	}

	return &GeneralTodoCategoryResponse{
		CategoryID: categoryID,
		Name:       req.Name,
		SortNumber: nextSort,
		Items:      []GeneralTodoItemResponse{},
	}, nil
}

func (s *GeneralTodoService) UpdateCategory(ctx context.Context, username string, categoryID int64, req UpdateCategoryRequest) error {
	memberID, err := s.getMemberID(ctx, username)
	if err != nil {
		return err
	}

	result, err := s.db.ExecContext(ctx,
		`UPDATE general_todo_category SET name = ?, last_modified_date = ?
		 WHERE general_todo_category_id = ? AND member_id = ?`,
		req.Name, time.Now(), categoryID, memberID,
	)
	if err != nil {
		return fmt.Errorf("update category: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("category not found or not owned by user")
	}
	return nil
}

func (s *GeneralTodoService) DeleteCategory(ctx context.Context, username string, categoryID int64) error {
	memberID, err := s.getMemberID(ctx, username)
	if err != nil {
		return err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// Delete statuses belonging to this category
	_, err = tx.ExecContext(ctx,
		`DELETE FROM general_todo_status WHERE category_id = ? AND member_id = ?`,
		categoryID, memberID,
	)
	if err != nil {
		return fmt.Errorf("delete statuses: %w", err)
	}

	// Delete items belonging to this category
	_, err = tx.ExecContext(ctx,
		`DELETE FROM general_todo_item WHERE category_id = ? AND member_id = ?`,
		categoryID, memberID,
	)
	if err != nil {
		return fmt.Errorf("delete items: %w", err)
	}

	// Delete the category itself
	result, err := tx.ExecContext(ctx,
		`DELETE FROM general_todo_category WHERE general_todo_category_id = ? AND member_id = ?`,
		categoryID, memberID,
	)
	if err != nil {
		return fmt.Errorf("delete category: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("category not found or not owned by user")
	}

	return tx.Commit()
}

func (s *GeneralTodoService) UpdateCategorySort(ctx context.Context, username string, items []CategorySortItem) error {
	memberID, err := s.getMemberID(ctx, username)
	if err != nil {
		return err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	now := time.Now()
	for _, item := range items {
		_, err := tx.ExecContext(ctx,
			`UPDATE general_todo_category SET sort_order = ?, last_modified_date = ?
			 WHERE general_todo_category_id = ? AND member_id = ?`,
			item.SortNumber, now, item.CategoryID, memberID,
		)
		if err != nil {
			return fmt.Errorf("update category sort: %w", err)
		}
	}

	return tx.Commit()
}

// --- Item ---

func (s *GeneralTodoService) CreateItem(ctx context.Context, username string, req CreateItemRequest) (*GeneralTodoItemResponse, error) {
	memberID, err := s.getMemberID(ctx, username)
	if err != nil {
		return nil, err
	}

	// Look up the folder_id from the category
	var folderID int64
	err = s.db.QueryRowContext(ctx,
		`SELECT folder_id FROM general_todo_category WHERE general_todo_category_id = ? AND member_id = ?`,
		req.CategoryID, memberID,
	).Scan(&folderID)
	if err != nil {
		return nil, fmt.Errorf("category not found: %w", err)
	}

	now := time.Now()
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO general_todo_item (folder_id, category_id, member_id, status_id, title, description, is_all_day, created_date, last_modified_date)
		 VALUES (?, ?, ?, 0, ?, NULL, false, ?, ?)`,
		folderID, req.CategoryID, memberID, req.Name, now, now,
	)
	if err != nil {
		return nil, fmt.Errorf("insert item: %w", err)
	}
	itemID, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("get last insert id: %w", err)
	}

	return &GeneralTodoItemResponse{
		ItemID:     itemID,
		Name:       req.Name,
		Memo:       "",
		SortNumber: 0,
		Statuses:   []GeneralTodoStatusResponse{},
	}, nil
}

func (s *GeneralTodoService) UpdateItem(ctx context.Context, username string, itemID int64, req UpdateItemRequest) error {
	memberID, err := s.getMemberID(ctx, username)
	if err != nil {
		return err
	}

	result, err := s.db.ExecContext(ctx,
		`UPDATE general_todo_item SET title = ?, description = ?, last_modified_date = ?
		 WHERE general_todo_item_id = ? AND member_id = ?`,
		req.Name, req.Memo, time.Now(), itemID, memberID,
	)
	if err != nil {
		return fmt.Errorf("update item: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("item not found or not owned by user")
	}
	return nil
}

func (s *GeneralTodoService) DeleteItem(ctx context.Context, username string, itemID int64) error {
	memberID, err := s.getMemberID(ctx, username)
	if err != nil {
		return err
	}

	result, err := s.db.ExecContext(ctx,
		`DELETE FROM general_todo_item WHERE general_todo_item_id = ? AND member_id = ?`,
		itemID, memberID,
	)
	if err != nil {
		return fmt.Errorf("delete item: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("item not found or not owned by user")
	}
	return nil
}

func (s *GeneralTodoService) UpdateItemSort(ctx context.Context, username string, items []ItemSortItem) error {
	memberID, err := s.getMemberID(ctx, username)
	if err != nil {
		return err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	now := time.Now()
	for _, item := range items {
		_, err := tx.ExecContext(ctx,
			`UPDATE general_todo_item SET status_id = ?, last_modified_date = ?
			 WHERE general_todo_item_id = ? AND member_id = ?`,
			item.SortNumber, now, item.ItemID, memberID,
		)
		if err != nil {
			return fmt.Errorf("update item sort: %w", err)
		}
	}

	return tx.Commit()
}

// --- Status ---

func (s *GeneralTodoService) CreateStatus(ctx context.Context, username string, req CreateStatusRequest) (*GeneralTodoStatusResponse, error) {
	memberID, err := s.getMemberID(ctx, username)
	if err != nil {
		return nil, err
	}

	// Use ItemID as the category_id reference for the status (per-item status tracking).
	// The name stores the character name, and sort_order 0 means unchecked.
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO general_todo_status (category_id, member_id, name, sort_order)
		 VALUES (?, ?, ?, 0)`,
		req.ItemID, memberID, req.CharacterName,
	)
	if err != nil {
		return nil, fmt.Errorf("insert status: %w", err)
	}
	statusID, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("get last insert id: %w", err)
	}

	return &GeneralTodoStatusResponse{
		StatusID:      statusID,
		CharacterName: req.CharacterName,
		Checked:       false,
	}, nil
}

func (s *GeneralTodoService) ToggleStatus(ctx context.Context, username string, statusID int64) error {
	memberID, err := s.getMemberID(ctx, username)
	if err != nil {
		return err
	}

	// Toggle: if sort_order is 0 set to 1, if 1 set to 0
	result, err := s.db.ExecContext(ctx,
		`UPDATE general_todo_status SET sort_order = 1 - sort_order
		 WHERE general_todo_status_id = ? AND member_id = ?`,
		statusID, memberID,
	)
	if err != nil {
		return fmt.Errorf("toggle status: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("status not found or not owned by user")
	}
	return nil
}

func (s *GeneralTodoService) DeleteStatus(ctx context.Context, username string, statusID int64) error {
	memberID, err := s.getMemberID(ctx, username)
	if err != nil {
		return err
	}

	result, err := s.db.ExecContext(ctx,
		`DELETE FROM general_todo_status WHERE general_todo_status_id = ? AND member_id = ?`,
		statusID, memberID,
	)
	if err != nil {
		return fmt.Errorf("delete status: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("status not found or not owned by user")
	}
	return nil
}
