package db

import (
	"context"
	"database/sql"
	"time"
)

// --- CommunityRepository ---

type CommunityRepository struct {
	db *sql.DB
}

func NewCommunityRepository(db *sql.DB) *CommunityRepository {
	return &CommunityRepository{db: db}
}

const communityColumns = `
	community_id, member_id, character_image, character_class_name, name,
	show_name, body, category, root_parent_id, comment_parent_id, deleted,
	created_date, last_modified_date
`

func scanCommunity(scanner interface{ Scan(dest ...interface{}) error }) (*Community, error) {
	c := &Community{}
	err := scanner.Scan(
		&c.ID, &c.MemberID, &c.CharacterImage, &c.CharacterClassName, &c.Name,
		&c.ShowName, &c.Body, &c.Category, &c.RootParentID, &c.CommentParentID, &c.Deleted,
		&c.CreatedDate, &c.LastModifiedDate,
	)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (r *CommunityRepository) FindByID(ctx context.Context, id int64) (*Community, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT `+communityColumns+` FROM community WHERE community_id = ?`, id,
	)
	return scanCommunity(row)
}

func (r *CommunityRepository) FindAll(ctx context.Context, offset, limit int) ([]Community, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+communityColumns+` FROM community
		 WHERE deleted = false AND root_parent_id = 0
		 ORDER BY created_date DESC
		 LIMIT ? OFFSET ?`, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var communities []Community
	for rows.Next() {
		c, err := scanCommunity(rows)
		if err != nil {
			return nil, err
		}
		communities = append(communities, *c)
	}
	return communities, rows.Err()
}

func (r *CommunityRepository) FindByCategory(ctx context.Context, category string, offset, limit int) ([]Community, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+communityColumns+` FROM community
		 WHERE deleted = false AND category = ? AND root_parent_id = 0
		 ORDER BY created_date DESC
		 LIMIT ? OFFSET ?`, category, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var communities []Community
	for rows.Next() {
		c, err := scanCommunity(rows)
		if err != nil {
			return nil, err
		}
		communities = append(communities, *c)
	}
	return communities, rows.Err()
}

func (r *CommunityRepository) FindReplies(ctx context.Context, rootParentID int64) ([]Community, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+communityColumns+` FROM community
		 WHERE root_parent_id = ? AND deleted = false
		 ORDER BY created_date ASC`, rootParentID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var communities []Community
	for rows.Next() {
		c, err := scanCommunity(rows)
		if err != nil {
			return nil, err
		}
		communities = append(communities, *c)
	}
	return communities, rows.Err()
}

func (r *CommunityRepository) Create(ctx context.Context, c *Community) (int64, error) {
	now := time.Now()
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO community (member_id, character_image, character_class_name, name,
		                        show_name, body, category, root_parent_id, comment_parent_id, deleted,
		                        created_date, last_modified_date)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		c.MemberID, c.CharacterImage, c.CharacterClassName, c.Name,
		c.ShowName, c.Body, c.Category, c.RootParentID, c.CommentParentID, false,
		now, now,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *CommunityRepository) UpdateBody(ctx context.Context, id int64, body string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE community SET body = ?, last_modified_date = ? WHERE community_id = ?`,
		body, time.Now(), id,
	)
	return err
}

func (r *CommunityRepository) SoftDelete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE community SET deleted = true, last_modified_date = ? WHERE community_id = ?`,
		time.Now(), id,
	)
	return err
}

func (r *CommunityRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM community WHERE community_id = ?`, id)
	return err
}

func (r *CommunityRepository) CountAll(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM community WHERE deleted = false AND root_parent_id = 0`,
	).Scan(&count)
	return count, err
}

func (r *CommunityRepository) CountByCategory(ctx context.Context, category string) (int64, error) {
	var count int64
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM community WHERE deleted = false AND category = ? AND root_parent_id = 0`,
		category,
	).Scan(&count)
	return count, err
}

func (r *CommunityRepository) CountReplies(ctx context.Context, rootParentID int64) (int64, error) {
	var count int64
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM community WHERE root_parent_id = ? AND deleted = false`,
		rootParentID,
	).Scan(&count)
	return count, err
}

// --- CommunityImagesRepository ---

type CommunityImagesRepository struct {
	db *sql.DB
}

func NewCommunityImagesRepository(db *sql.DB) *CommunityImagesRepository {
	return &CommunityImagesRepository{db: db}
}

func (r *CommunityImagesRepository) FindByCommunityID(ctx context.Context, communityID int64) ([]CommunityImages, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT community_images_id, community_id, url, file_name, ordering, deleted,
		        created_date, last_modified_date
		 FROM community_images WHERE community_id = ? AND deleted = false ORDER BY ordering ASC`,
		communityID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var images []CommunityImages
	for rows.Next() {
		var img CommunityImages
		if err := rows.Scan(
			&img.ID, &img.CommunityID, &img.URL, &img.FileName, &img.Ordering, &img.Deleted,
			&img.CreatedDate, &img.LastModifiedDate,
		); err != nil {
			return nil, err
		}
		images = append(images, img)
	}
	return images, rows.Err()
}

func (r *CommunityImagesRepository) Create(ctx context.Context, img *CommunityImages) (int64, error) {
	now := time.Now()
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO community_images (community_id, url, file_name, ordering, deleted,
		                               created_date, last_modified_date)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		img.CommunityID, img.URL, img.FileName, img.Ordering, false,
		now, now,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *CommunityImagesRepository) SoftDelete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE community_images SET deleted = true, last_modified_date = ? WHERE community_images_id = ?`,
		time.Now(), id,
	)
	return err
}

func (r *CommunityImagesRepository) UpdateOrdering(ctx context.Context, id int64, communityID int64, ordering int) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE community_images SET community_id = ?, ordering = ?, last_modified_date = ?
		 WHERE community_images_id = ?`,
		communityID, ordering, time.Now(), id,
	)
	return err
}

// --- CommunityLikeRepository ---

type CommunityLikeRepository struct {
	db *sql.DB
}

func NewCommunityLikeRepository(db *sql.DB) *CommunityLikeRepository {
	return &CommunityLikeRepository{db: db}
}

func (r *CommunityLikeRepository) FindByCommunityIDAndMemberID(ctx context.Context, communityID, memberID int64) (*CommunityLike, error) {
	cl := &CommunityLike{}
	err := r.db.QueryRowContext(ctx,
		`SELECT community_like_id, member_id, community_id, created_date, last_modified_date
		 FROM community_like WHERE community_id = ? AND member_id = ?`,
		communityID, memberID,
	).Scan(&cl.ID, &cl.MemberID, &cl.CommunityID, &cl.CreatedDate, &cl.LastModifiedDate)
	if err != nil {
		return nil, err
	}
	return cl, nil
}

func (r *CommunityLikeRepository) Create(ctx context.Context, communityID, memberID int64) (int64, error) {
	now := time.Now()
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO community_like (community_id, member_id, created_date, last_modified_date)
		 VALUES (?, ?, ?, ?)`,
		communityID, memberID, now, now,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *CommunityLikeRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM community_like WHERE community_like_id = ?`, id)
	return err
}

func (r *CommunityLikeRepository) CountByCommunityID(ctx context.Context, communityID int64) (int64, error) {
	var count int64
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM community_like WHERE community_id = ?`, communityID,
	).Scan(&count)
	return count, err
}

// --- CommentsRepository ---

type CommentsRepository struct {
	db *sql.DB
}

func NewCommentsRepository(db *sql.DB) *CommentsRepository {
	return &CommentsRepository{db: db}
}

func (r *CommentsRepository) FindByID(ctx context.Context, id int64) (*Comments, error) {
	c := &Comments{}
	err := r.db.QueryRowContext(ctx,
		`SELECT comments_id, body, member_id, parent_id, created_date, last_modified_date
		 FROM comments WHERE comments_id = ?`, id,
	).Scan(&c.ID, &c.Body, &c.MemberID, &c.ParentID, &c.CreatedDate, &c.LastModifiedDate)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (r *CommentsRepository) FindByParentID(ctx context.Context, parentID int64) ([]Comments, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT comments_id, body, member_id, parent_id, created_date, last_modified_date
		 FROM comments WHERE parent_id = ? ORDER BY created_date ASC`, parentID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []Comments
	for rows.Next() {
		var c Comments
		if err := rows.Scan(&c.ID, &c.Body, &c.MemberID, &c.ParentID, &c.CreatedDate, &c.LastModifiedDate); err != nil {
			return nil, err
		}
		comments = append(comments, c)
	}
	return comments, rows.Err()
}

func (r *CommentsRepository) Create(ctx context.Context, c *Comments) (int64, error) {
	now := time.Now()
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO comments (body, member_id, parent_id, created_date, last_modified_date)
		 VALUES (?, ?, ?, ?, ?)`,
		c.Body, c.MemberID, c.ParentID, now, now,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *CommentsRepository) UpdateBody(ctx context.Context, id int64, body string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE comments SET body = ?, last_modified_date = ? WHERE comments_id = ?`,
		body, time.Now(), id,
	)
	return err
}

func (r *CommentsRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM comments WHERE comments_id = ?`, id)
	return err
}

// --- FollowRepository ---

type FollowRepository struct {
	db *sql.DB
}

func NewFollowRepository(db *sql.DB) *FollowRepository {
	return &FollowRepository{db: db}
}

func (r *FollowRepository) FindByFollowerIDAndFollowingID(ctx context.Context, followerID, followingID int64) (*Follow, error) {
	f := &Follow{}
	err := r.db.QueryRowContext(ctx,
		`SELECT follow_id, follower_id, following_id FROM follow
		 WHERE follower_id = ? AND following_id = ?`,
		followerID, followingID,
	).Scan(&f.ID, &f.FollowerID, &f.FollowingID)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func (r *FollowRepository) FindFollowingsByFollowerID(ctx context.Context, followerID int64) ([]Follow, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT follow_id, follower_id, following_id FROM follow WHERE follower_id = ?`, followerID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var follows []Follow
	for rows.Next() {
		var f Follow
		if err := rows.Scan(&f.ID, &f.FollowerID, &f.FollowingID); err != nil {
			return nil, err
		}
		follows = append(follows, f)
	}
	return follows, rows.Err()
}

func (r *FollowRepository) FindFollowersByFollowingID(ctx context.Context, followingID int64) ([]Follow, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT follow_id, follower_id, following_id FROM follow WHERE following_id = ?`, followingID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var follows []Follow
	for rows.Next() {
		var f Follow
		if err := rows.Scan(&f.ID, &f.FollowerID, &f.FollowingID); err != nil {
			return nil, err
		}
		follows = append(follows, f)
	}
	return follows, rows.Err()
}

func (r *FollowRepository) Create(ctx context.Context, followerID, followingID int64) (int64, error) {
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO follow (follower_id, following_id) VALUES (?, ?)`,
		followerID, followingID,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *FollowRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM follow WHERE follow_id = ?`, id)
	return err
}

func (r *FollowRepository) DeleteByFollowerIDAndFollowingID(ctx context.Context, followerID, followingID int64) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM follow WHERE follower_id = ? AND following_id = ?`,
		followerID, followingID,
	)
	return err
}

func (r *FollowRepository) CountFollowers(ctx context.Context, followingID int64) (int64, error) {
	var count int64
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM follow WHERE following_id = ?`, followingID,
	).Scan(&count)
	return count, err
}

func (r *FollowRepository) CountFollowings(ctx context.Context, followerID int64) (int64, error) {
	var count int64
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM follow WHERE follower_id = ?`, followerID,
	).Scan(&count)
	return count, err
}
