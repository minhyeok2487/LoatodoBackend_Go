package service

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type CommentsService struct {
	db *sql.DB
}

func NewCommentsService(db *sql.DB) *CommentsService {
	return &CommentsService{db: db}
}

type CommentResponse struct {
	CommentID   int64              `json:"commentId"`
	CommunityID int64             `json:"communityId"`
	Username    string             `json:"username"`
	Body        string             `json:"body"`
	ParentID    *int64             `json:"parentId"`
	Children    []CommentResponse  `json:"children"`
	CreatedAt   time.Time          `json:"createdAt"`
	UpdatedAt   time.Time          `json:"updatedAt"`
}

type CreateCommentRequest struct {
	CommunityID int64  `json:"communityId"`
	Body        string `json:"body"`
	ParentID    *int64 `json:"parentId"`
}

type UpdateCommentRequest struct {
	Body string `json:"body"`
}

// getMemberID resolves the member_id for a given username.
func (s *CommentsService) getMemberID(ctx context.Context, username string) (int64, error) {
	var memberID int64
	err := s.db.QueryRowContext(ctx,
		`SELECT member_id FROM member WHERE username = ?`, username,
	).Scan(&memberID)
	if err != nil {
		return 0, fmt.Errorf("member not found: %w", err)
	}
	return memberID, nil
}

// getUsername resolves the username for a given member_id.
func (s *CommentsService) getUsername(ctx context.Context, memberID int64) string {
	var username sql.NullString
	_ = s.db.QueryRowContext(ctx,
		`SELECT username FROM member WHERE member_id = ?`, memberID,
	).Scan(&username)
	if username.Valid {
		return username.String
	}
	return ""
}

func (s *CommentsService) GetComments(ctx context.Context, communityID int64) ([]CommentResponse, error) {
	// Fetch all comments for this community post.
	// Comments are linked via root_parent_id in the community table, but since we
	// have a dedicated comments table, we query comments whose parent chain
	// references this communityID.
	// The comments table stores: comments_id, body, member_id, parent_id, created_date, last_modified_date.
	// parent_id = 0 means it's a root comment for the community post.
	// We use the community_id context: root comments have parent_id = communityID's comments_id,
	// but actually parent_id=0 means root-level.
	//
	// Since there's no community_id column in the comments table, the linkage is:
	// root comments have parent_id = <communityID> (the community post ID acts as the parent).
	// Nested comments have parent_id = <another comment's ID>.
	//
	// Load all comments where the root references the community post.
	// Strategy: load all comments that are either direct children of communityID
	// or children of those comments (recursive tree). For simplicity, we load in two passes:
	// 1) root comments: parent_id = communityID
	// 2) all their children via parent_id matching any root comment ID.
	// For deeper nesting, we do a broader approach: load all comments with parent_id = communityID
	// and all comments whose parent_id is in the set of already-loaded IDs.

	// Actually, simpler approach: load all comments where parent_id = communityID (root level)
	// and recursively all replies. Since this is a flat table, load all comments whose
	// root parent chain starts at communityID.
	// Simplest: load all root comments (parent_id = communityID), then load all others
	// whose parent_id is any loaded comment. Repeat until no new ones found.

	// For efficiency with typical shallow trees, load root + one level of replies:
	rootRows, err := s.db.QueryContext(ctx,
		`SELECT comments_id, COALESCE(body, ''), member_id, parent_id, created_date, last_modified_date
		 FROM comments WHERE parent_id = ? ORDER BY created_date ASC`,
		communityID,
	)
	if err != nil {
		return nil, fmt.Errorf("query root comments: %w", err)
	}
	defer rootRows.Close()

	type commentRow struct {
		id        int64
		body      string
		memberID  int64
		parentID  int64
		created   time.Time
		modified  time.Time
	}
	var roots []commentRow
	var rootIDs []int64
	for rootRows.Next() {
		var c commentRow
		if err := rootRows.Scan(&c.id, &c.body, &c.memberID, &c.parentID, &c.created, &c.modified); err != nil {
			return nil, err
		}
		roots = append(roots, c)
		rootIDs = append(rootIDs, c.id)
	}
	if err := rootRows.Err(); err != nil {
		return nil, err
	}

	// Load all child comments (replies to root comments)
	childMap := make(map[int64][]commentRow)
	if len(rootIDs) > 0 {
		query := `SELECT comments_id, COALESCE(body, ''), member_id, parent_id, created_date, last_modified_date
		          FROM comments WHERE parent_id IN (`
		args := make([]interface{}, len(rootIDs))
		for i, id := range rootIDs {
			if i > 0 {
				query += ","
			}
			query += "?"
			args[i] = id
		}
		query += ") ORDER BY created_date ASC"

		childRows, err := s.db.QueryContext(ctx, query, args...)
		if err != nil {
			return nil, fmt.Errorf("query child comments: %w", err)
		}
		defer childRows.Close()

		for childRows.Next() {
			var c commentRow
			if err := childRows.Scan(&c.id, &c.body, &c.memberID, &c.parentID, &c.created, &c.modified); err != nil {
				return nil, err
			}
			childMap[c.parentID] = append(childMap[c.parentID], c)
		}
		if err := childRows.Err(); err != nil {
			return nil, err
		}
	}

	// Build response tree
	result := make([]CommentResponse, 0, len(roots))
	for _, r := range roots {
		parentID := r.parentID
		var pID *int64
		if parentID != 0 {
			pID = &parentID
		}

		children := make([]CommentResponse, 0)
		for _, ch := range childMap[r.id] {
			chParent := ch.parentID
			children = append(children, CommentResponse{
				CommentID:   ch.id,
				CommunityID: communityID,
				Username:    s.getUsername(ctx, ch.memberID),
				Body:        ch.body,
				ParentID:    &chParent,
				Children:    []CommentResponse{},
				CreatedAt:   ch.created,
				UpdatedAt:   ch.modified,
			})
		}

		result = append(result, CommentResponse{
			CommentID:   r.id,
			CommunityID: communityID,
			Username:    s.getUsername(ctx, r.memberID),
			Body:        r.body,
			ParentID:    pID,
			Children:    children,
			CreatedAt:   r.created,
			UpdatedAt:   r.modified,
		})
	}

	return result, nil
}

func (s *CommentsService) CreateComment(ctx context.Context, username string, req CreateCommentRequest) (*CommentResponse, error) {
	memberID, err := s.getMemberID(ctx, username)
	if err != nil {
		return nil, err
	}

	parentID := int64(0)
	if req.ParentID != nil {
		parentID = *req.ParentID
	} else {
		// Root-level comment: parent_id = communityID
		parentID = req.CommunityID
	}

	now := time.Now()
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO comments (body, member_id, parent_id, created_date, last_modified_date)
		 VALUES (?, ?, ?, ?, ?)`,
		req.Body, memberID, parentID, now, now,
	)
	if err != nil {
		return nil, fmt.Errorf("insert comment: %w", err)
	}
	commentID, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("get last insert id: %w", err)
	}

	var pID *int64
	if req.ParentID != nil {
		pID = req.ParentID
	}

	return &CommentResponse{
		CommentID:   commentID,
		CommunityID: req.CommunityID,
		Username:    username,
		Body:        req.Body,
		ParentID:    pID,
		Children:    []CommentResponse{},
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

func (s *CommentsService) UpdateComment(ctx context.Context, username string, commentID int64, req UpdateCommentRequest) error {
	memberID, err := s.getMemberID(ctx, username)
	if err != nil {
		return err
	}

	result, err := s.db.ExecContext(ctx,
		`UPDATE comments SET body = ?, last_modified_date = ?
		 WHERE comments_id = ? AND member_id = ?`,
		req.Body, time.Now(), commentID, memberID,
	)
	if err != nil {
		return fmt.Errorf("update comment: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("comment not found or not owned by user")
	}
	return nil
}

func (s *CommentsService) DeleteComment(ctx context.Context, username string, commentID int64) error {
	memberID, err := s.getMemberID(ctx, username)
	if err != nil {
		return err
	}

	result, err := s.db.ExecContext(ctx,
		`DELETE FROM comments WHERE comments_id = ? AND member_id = ?`,
		commentID, memberID,
	)
	if err != nil {
		return fmt.Errorf("delete comment: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("comment not found or not owned by user")
	}
	return nil
}
