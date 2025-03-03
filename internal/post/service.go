package post

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/thediligencedev/betteridn/internal/models"
)

var (
	ErrPostNotFound      = errors.New("post not found")
	ErrCategoryNotFound  = errors.New("one or more categories not found")
	ErrUnauthorized      = errors.New("unauthorized to modify this post")
	ErrInvalidVoteType   = errors.New("invalid vote type, must be 1 (upvote) or -1 (downvote)")
	ErrValidationFailed  = errors.New("validation failed")
	ErrInternalServer    = errors.New("internal server error")
	ErrDuplicateVote     = errors.New("user has already voted on this post")
	ErrInvalidPagination = errors.New("invalid pagination parameters")
)

type PostService struct {
	pool *pgxpool.Pool
}

func NewPostService(pool *pgxpool.Pool) *PostService {
	return &PostService{pool: pool}
}

// CreatePost creates a new post with the given title, content, and categories
func (s *PostService) CreatePost(ctx context.Context, userID uuid.UUID, title, content string, categories []string) (uuid.UUID, error) {
	// Validate categories exist
	if err := s.validateCategories(ctx, categories); err != nil {
		return uuid.Nil, err
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return uuid.Nil, ErrInternalServer
	}
	defer tx.Rollback(ctx)

	// Insert post
	var postID uuid.UUID
	err = tx.QueryRow(ctx, `
		INSERT INTO posts (user_id, title, content, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())
		RETURNING id
	`, userID, title, content).Scan(&postID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to create post: %w", err)
	}

	// Insert post categories
	for _, categoryName := range categories {
		var categoryID uuid.UUID
		err = tx.QueryRow(ctx, `
			SELECT id FROM categories WHERE name = $1
		`, categoryName).Scan(&categoryID)
		if err != nil {
			return uuid.Nil, ErrCategoryNotFound
		}

		_, err = tx.Exec(ctx, `
			INSERT INTO post_categories (post_id, category_id)
			VALUES ($1, $2)
		`, postID, categoryID)
		if err != nil {
			return uuid.Nil, fmt.Errorf("failed to associate post with category: %w", err)
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return uuid.Nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return postID, nil
}

// GetPosts retrieves a paginated list of posts
func (s *PostService) GetPosts(ctx context.Context, page, limit int) ([]models.Post, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	// Query posts with user info
	rows, err := s.pool.Query(ctx, `
		SELECT p.id, p.title, p.content, p.created_at, p.updated_at, u.username
		FROM posts p
		JOIN users u ON p.user_id = u.id
		ORDER BY p.created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query posts: %w", err)
	}
	defer rows.Close()

	var posts []models.Post
	for rows.Next() {
		var post models.Post
		var username string

		err := rows.Scan(
			&post.ID,
			&post.Title,
			&post.Content,
			&post.CreatedAt,
			&post.UpdatedAt,
			&username,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan post row: %w", err)
		}

		post.User = &models.UserBasic{
			Username: username,
		}

		// Get categories for this post
		post.Categories, err = s.getPostCategories(ctx, post.ID)
		if err != nil {
			return nil, err
		}

		// Get vote counts for this post
		post.VoteCount, err = s.getPostVoteCounts(ctx, post.ID)
		if err != nil {
			return nil, err
		}

		posts = append(posts, post)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating post rows: %w", err)
	}

	return posts, nil
}

// GetPostByID retrieves a post by its ID
func (s *PostService) GetPostByID(ctx context.Context, postID uuid.UUID) (*models.Post, error) {
	var post models.Post
	var username string

	err := s.pool.QueryRow(ctx, `
		SELECT p.id, p.title, p.content, p.created_at, p.updated_at, u.username
		FROM posts p
		JOIN users u ON p.user_id = u.id
		WHERE p.id = $1
	`, postID).Scan(
		&post.ID,
		&post.Title,
		&post.Content,
		&post.CreatedAt,
		&post.UpdatedAt,
		&username,
	)
	if err != nil {
		return nil, ErrPostNotFound
	}

	post.User = &models.UserBasic{
		Username: username,
	}

	// Get categories for this post
	post.Categories, err = s.getPostCategories(ctx, post.ID)
	if err != nil {
		return nil, err
	}

	// Get vote counts for this post
	post.VoteCount, err = s.getPostVoteCounts(ctx, post.ID)
	if err != nil {
		return nil, err
	}

	return &post, nil
}

// UpdatePost updates an existing post
func (s *PostService) UpdatePost(ctx context.Context, postID, userID uuid.UUID, title, content string, categories []string) error {
	// Check if post exists and belongs to the user
	var postOwnerID uuid.UUID
	err := s.pool.QueryRow(ctx, `
		SELECT user_id FROM posts WHERE id = $1
	`, postID).Scan(&postOwnerID)
	if err != nil {
		return ErrPostNotFound
	}

	if postOwnerID != userID {
		return ErrUnauthorized
	}

	// Validate categories exist
	if err := s.validateCategories(ctx, categories); err != nil {
		return err
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return ErrInternalServer
	}
	defer tx.Rollback(ctx)

	// Update post
	_, err = tx.Exec(ctx, `
		UPDATE posts
		SET title = $1, content = $2, updated_at = $3
		WHERE id = $4
	`, title, content, time.Now(), postID)
	if err != nil {
		return fmt.Errorf("failed to update post: %w", err)
	}

	// Delete existing post categories
	_, err = tx.Exec(ctx, `
		DELETE FROM post_categories
		WHERE post_id = $1
	`, postID)
	if err != nil {
		return fmt.Errorf("failed to delete existing post categories: %w", err)
	}

	// Insert new post categories
	for _, categoryName := range categories {
		var categoryID uuid.UUID
		err = tx.QueryRow(ctx, `
			SELECT id FROM categories WHERE name = $1
		`, categoryName).Scan(&categoryID)
		if err != nil {
			return ErrCategoryNotFound
		}

		_, err = tx.Exec(ctx, `
			INSERT INTO post_categories (post_id, category_id)
			VALUES ($1, $2)
		`, postID, categoryID)
		if err != nil {
			return fmt.Errorf("failed to associate post with category: %w", err)
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// VotePost records a vote on a post
func (s *PostService) VotePost(ctx context.Context, postID, userID uuid.UUID, voteType int) (*models.VoteResult, error) {
	// Validate vote type
	if voteType != 1 && voteType != -1 {
		return nil, ErrInvalidVoteType
	}

	// Check if post exists
	var exists bool
	err := s.pool.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM posts WHERE id = $1)
	`, postID).Scan(&exists)
	if err != nil {
		return nil, ErrInternalServer
	}
	if !exists {
		return nil, ErrPostNotFound
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, ErrInternalServer
	}
	defer tx.Rollback(ctx)

	// Track if the vote was removed
	voteRemoved := false

	// Check if user has already voted on this post
	var existingVoteID uuid.UUID
	var existingVoteType int
	err = tx.QueryRow(ctx, `
		SELECT id, vote_type FROM post_votes
		WHERE post_id = $1 AND user_id = $2
	`, postID, userID).Scan(&existingVoteID, &existingVoteType)

	if err == nil {
		// User has already voted, update the vote
		if existingVoteType == voteType {
			// Same vote type, remove the vote
			_, err = tx.Exec(ctx, `
				DELETE FROM post_votes
				WHERE id = $1
			`, existingVoteID)
			if err != nil {
				return nil, fmt.Errorf("failed to remove vote: %w", err)
			}
			voteRemoved = true
		} else {
			// Different vote type, update the vote
			_, err = tx.Exec(ctx, `
				UPDATE post_votes
				SET vote_type = $1
				WHERE id = $2
			`, voteType, existingVoteID)
			if err != nil {
				return nil, fmt.Errorf("failed to update vote: %w", err)
			}
		}
	} else {
		// User has not voted yet, insert new vote
		_, err = tx.Exec(ctx, `
			INSERT INTO post_votes (post_id, user_id, vote_type)
			VALUES ($1, $2, $3)
		`, postID, userID, voteType)
		if err != nil {
			return nil, fmt.Errorf("failed to insert vote: %w", err)
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Get updated vote counts
	voteCount, err := s.getPostVoteCounts(ctx, postID)
	if err != nil {
		return nil, err
	}

	return &models.VoteResult{
		VoteCount:   *voteCount,
		VoteRemoved: voteRemoved,
	}, nil
}

// Helper functions

// validateCategories checks if all categories exist
func (s *PostService) validateCategories(ctx context.Context, categories []string) error {
	if len(categories) == 0 {
		return ErrValidationFailed
	}

	for _, category := range categories {
		var exists bool
		err := s.pool.QueryRow(ctx, `
			SELECT EXISTS(SELECT 1 FROM categories WHERE name = $1)
		`, category).Scan(&exists)
		if err != nil {
			return ErrInternalServer
		}
		if !exists {
			return ErrCategoryNotFound
		}
	}
	return nil
}

// getPostCategories retrieves categories for a post
func (s *PostService) getPostCategories(ctx context.Context, postID uuid.UUID) ([]string, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT c.name
		FROM categories c
		JOIN post_categories pc ON c.id = pc.category_id
		WHERE pc.post_id = $1
	`, postID)
	if err != nil {
		return nil, fmt.Errorf("failed to query post categories: %w", err)
	}
	defer rows.Close()

	var categories []string
	for rows.Next() {
		var category string
		if err := rows.Scan(&category); err != nil {
			return nil, fmt.Errorf("failed to scan category: %w", err)
		}
		categories = append(categories, category)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating category rows: %w", err)
	}

	return categories, nil
}

// getPostVoteCounts retrieves vote counts for a post
func (s *PostService) getPostVoteCounts(ctx context.Context, postID uuid.UUID) (*models.VoteCount, error) {
	var upvotes, downvotes int

	err := s.pool.QueryRow(ctx, `
		SELECT 
			COUNT(CASE WHEN vote_type = 1 THEN 1 END) as upvotes,
			COUNT(CASE WHEN vote_type = -1 THEN 1 END) as downvotes
		FROM post_votes
		WHERE post_id = $1
	`, postID).Scan(&upvotes, &downvotes)

	if err != nil {
		return nil, fmt.Errorf("failed to get vote counts: %w", err)
	}

	return &models.VoteCount{
		Upvotes:   upvotes,
		Downvotes: downvotes,
	}, nil
}
