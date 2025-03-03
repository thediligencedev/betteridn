package post

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/thediligencedev/betteridn/internal/models"
	"github.com/thediligencedev/betteridn/pkg/response"
	"github.com/thediligencedev/betteridn/pkg/validator"
)

type Handler struct {
	service *PostService
}

func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{
		service: NewPostService(pool),
	}
}

// TODO: if still error, change categories to type interface{} so can get both array and string
type CreatePostRequest struct {
	Title      string   `json:"title" validate:"required"`
	Content    string   `json:"content" validate:"required"`
	Categories []string `json:"categories" validate:"required,min=1"`
}

type UpdatePostRequest struct {
	Title      string   `json:"title" validate:"required"`
	Content    string   `json:"content" validate:"required"`
	Categories []string `json:"categories" validate:"required,min=1"`
}

type VotePostRequest struct {
	VoteType int `json:"vote_type" validate:"required,oneof=1 -1"`
}

// CreatePost handles the creation of a new post
func (h *Handler) CreatePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.RespondWithError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// println(r.Context().Value(models.UserContextKey))

	// Get user ID from context
	userIDRaw := r.Context().Value(models.UserContextKey)
	if userIDRaw == nil {
		response.RespondWithError(w, http.StatusUnauthorized, "unauthorized: user_id not found in session")
		return
	}

	userID, ok := userIDRaw.(string)
	if !ok || userID == "" {
		response.RespondWithError(w, http.StatusUnauthorized, "unauthorized: invalid user_id in session")
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		response.RespondWithError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	// Parse request body
	var req CreatePostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.RespondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate request
	if err := validator.ValidateStruct(&req); err != nil {
		response.RespondWithError(w, http.StatusBadRequest, "validation error: "+err.Error())
		return
	}

	// Create post
	postID, err := h.service.CreatePost(r.Context(), userUUID, req.Title, req.Content, req.Categories)
	if err != nil {
		switch err {
		case ErrCategoryNotFound:
			response.RespondWithError(w, http.StatusBadRequest, "one or more categories not found")
		case ErrValidationFailed:
			response.RespondWithError(w, http.StatusBadRequest, "validation failed")
		default:
			log.Printf("CreatePost error: %v", err)
			response.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	// Success
	responseJSON := map[string]interface{}{
		"message": "post created successfully",
		"data": map[string]string{
			"id": postID.String(),
		},
	}
	response.RespondWithJSON(w, http.StatusCreated, responseJSON)
}

// GetPosts handles retrieving a paginated list of posts
func (h *Handler) GetPosts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.RespondWithError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// Parse pagination parameters
	page := 1
	limit := 20

	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if pageNum, err := strconv.Atoi(pageStr); err == nil && pageNum > 0 {
			page = pageNum
		}
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limitNum, err := strconv.Atoi(limitStr); err == nil && limitNum > 0 && limitNum <= 100 {
			limit = limitNum
		}
	}

	// Get posts
	posts, err := h.service.GetPosts(r.Context(), page, limit)
	if err != nil {
		log.Printf("GetPosts error: %v", err)
		response.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	// Success
	responseJSON := map[string]interface{}{
		"message": "posts retrieved successfully",
		"data":    posts,
	}
	response.RespondWithJSON(w, http.StatusOK, responseJSON)
}

// GetPostByID handles retrieving a single post by ID
func (h *Handler) GetPostByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.RespondWithError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// Extract post ID from URL path
	// Assuming the URL pattern is /api/v1/posts/{postId}
	path := r.URL.Path
	parts := []rune(path)
	lastSlashIndex := -1
	for i := len(parts) - 1; i >= 0; i-- {
		if parts[i] == '/' {
			lastSlashIndex = i
			break
		}
	}
	if lastSlashIndex == -1 || lastSlashIndex == len(parts)-1 {
		response.RespondWithError(w, http.StatusBadRequest, "invalid post ID")
		return
	}
	postIDStr := string(parts[lastSlashIndex+1:])

	postID, err := uuid.Parse(postIDStr)
	if err != nil {
		response.RespondWithError(w, http.StatusBadRequest, "invalid post ID format")
		return
	}

	// Get post
	post, err := h.service.GetPostByID(r.Context(), postID)
	if err != nil {
		switch err {
		case ErrPostNotFound:
			response.RespondWithError(w, http.StatusNotFound, "post not found")
		default:
			log.Printf("GetPostByID error: %v", err)
			response.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	// Success
	responseJSON := map[string]interface{}{
		"message": "post retrieved successfully",
		"data":    post,
	}
	response.RespondWithJSON(w, http.StatusOK, responseJSON)
}

// UpdatePost handles updating an existing post
func (h *Handler) UpdatePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		response.RespondWithError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// Get user ID from context
	userID := r.Context().Value(models.UserContextKey).(string)
	if userID == "" {
		response.RespondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		response.RespondWithError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	// Extract post ID from URL path
	path := r.URL.Path
	parts := []rune(path)
	lastSlashIndex := -1
	for i := len(parts) - 1; i >= 0; i-- {
		if parts[i] == '/' {
			lastSlashIndex = i
			break
		}
	}
	if lastSlashIndex == -1 || lastSlashIndex == len(parts)-1 {
		response.RespondWithError(w, http.StatusBadRequest, "invalid post ID")
		return
	}
	postIDStr := string(parts[lastSlashIndex+1:])

	postID, err := uuid.Parse(postIDStr)
	if err != nil {
		response.RespondWithError(w, http.StatusBadRequest, "invalid post ID format")
		return
	}

	// Parse request body
	var req UpdatePostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.RespondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate request
	if err := validator.ValidateStruct(&req); err != nil {
		response.RespondWithError(w, http.StatusBadRequest, "validation error: "+err.Error())
		return
	}

	// Update post
	err = h.service.UpdatePost(r.Context(), postID, userUUID, req.Title, req.Content, req.Categories)
	if err != nil {
		switch err {
		case ErrPostNotFound:
			response.RespondWithError(w, http.StatusNotFound, "post not found")
		case ErrUnauthorized:
			response.RespondWithError(w, http.StatusForbidden, "you are not authorized to update this post")
		case ErrCategoryNotFound:
			response.RespondWithError(w, http.StatusBadRequest, "one or more categories not found")
		case ErrValidationFailed:
			response.RespondWithError(w, http.StatusBadRequest, "validation failed")
		default:
			log.Printf("UpdatePost error: %v", err)
			response.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	// Success
	responseJSON := map[string]interface{}{
		"message": "post updated successfully",
		"data": map[string]string{
			"id": postID.String(),
		},
	}
	response.RespondWithJSON(w, http.StatusOK, responseJSON)
}

// VotePost handles voting on a post
func (h *Handler) VotePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.RespondWithError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// Get user ID from context
	userID := r.Context().Value(models.UserContextKey).(string)
	if userID == "" {
		response.RespondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		response.RespondWithError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	// Extract post ID from URL path
	// Assuming the URL pattern is /api/v1/posts/{postId}/vote
	path := r.URL.Path
	parts := []rune(path)

	// Find the second-to-last slash (before /vote)
	lastSlashIndex := -1
	secondLastSlashIndex := -1
	slashCount := 0

	for i := len(parts) - 1; i >= 0; i-- {
		if parts[i] == '/' {
			slashCount++
			if slashCount == 1 {
				lastSlashIndex = i
			} else if slashCount == 2 {
				secondLastSlashIndex = i
				break
			}
		}
	}

	if secondLastSlashIndex == -1 || lastSlashIndex == -1 {
		response.RespondWithError(w, http.StatusBadRequest, "invalid URL format")
		return
	}

	postIDStr := string(parts[secondLastSlashIndex+1 : lastSlashIndex])

	postID, err := uuid.Parse(postIDStr)
	if err != nil {
		response.RespondWithError(w, http.StatusBadRequest, "invalid post ID format")
		return
	}

	// Parse request body
	var req VotePostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.RespondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate request
	if err := validator.ValidateStruct(&req); err != nil {
		response.RespondWithError(w, http.StatusBadRequest, "validation error: "+err.Error())
		return
	}

	// Vote on post
	voteCount, err := h.service.VotePost(r.Context(), postID, userUUID, req.VoteType)
	if err != nil {
		switch err {
		case ErrPostNotFound:
			response.RespondWithError(w, http.StatusNotFound, "post not found")
		case ErrInvalidVoteType:
			response.RespondWithError(w, http.StatusBadRequest, "invalid vote type, must be 1 (upvote) or -1 (downvote)")
		default:
			log.Printf("VotePost error: %v", err)
			response.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	// Success
	responseJSON := map[string]interface{}{
		"message": "vote recorded successfully",
		"data": map[string]interface{}{
			"id":         postID.String(),
			"vote_count": voteCount,
		},
	}
	response.RespondWithJSON(w, http.StatusOK, responseJSON)
}
