package auth

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/alexedwards/scs/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/thediligencedev/betteridn/pkg/response"
	"github.com/thediligencedev/betteridn/pkg/validator"
)

// Handler holds dependencies for auth handlers.
type Handler struct {
	service        *AuthService
	sessionManager *scs.SessionManager
}

// NewHandler creates a new Handler.
func NewHandler(pool *pgxpool.Pool, sessionManager *scs.SessionManager) *Handler {
	return &Handler{
		service:        NewAuthService(pool),
		sessionManager: sessionManager,
	}
}

// SignUpRequest is the expected request body for sign up.
type SignUpRequest struct {
	Username string `json:"username" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

// SignInRequest is the expected request body for sign in.
type SignInRequest struct {
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// SignUp creates a new user account
func (h *Handler) SignUp(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.RespondWithError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req SignUpRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.RespondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := validator.ValidateStruct(&req); err != nil {
		response.RespondWithError(w, http.StatusBadRequest, "validation error")
		return
	}

	err := h.service.SignUp(r.Context(), req.Username, req.Email, req.Password)
	if err != nil {
		switch err {
		case ErrUserAlreadyExists:
			response.RespondWithError(w, http.StatusConflict, "user already exists")
			return
		case ErrCreateUser:
			response.RespondWithError(w, http.StatusInternalServerError, "internal server error")
			return
		default:
			log.Printf("SignUp error: %v", err)
			response.RespondWithError(w, http.StatusInternalServerError, "internal server error")
			return
		}
	}

	// Success response
	responseJSON := map[string]string{"message": "successfully created user"}
	response.RespondWithJSON(w, http.StatusOK, responseJSON)
}

// SignIn logs a user in (creates or renews a session)
func (h *Handler) SignIn(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.RespondWithError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req SignInRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.RespondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := validator.ValidateStruct(&req); err != nil {
		response.RespondWithError(w, http.StatusBadRequest, "validation error")
		return
	}

	ctx := r.Context()

	// Check if user is already logged in
	currentUserID := h.sessionManager.GetString(ctx, "user_id")
	if currentUserID != "" {
		// Already logged in -> renew token
		err := h.sessionManager.RenewToken(ctx)
		if err != nil {
			log.Printf("Failed to renew session token: %v", err)
			response.RespondWithError(w, http.StatusInternalServerError, "internal server error")
			return
		}

		// Return success with current user_id
		responseJSON := map[string]interface{}{
			"message": "user successfully signed in",
			"data": map[string]string{
				"user_id": currentUserID,
			},
		}
		response.RespondWithJSON(w, http.StatusOK, responseJSON)
		return
	}

	// Not logged in -> check credentials
	userID, err := h.service.SignIn(ctx, req.Email, req.Password)
	if err != nil {
		switch err {
		case ErrInvalidCredentials:
			response.RespondWithError(w, http.StatusUnauthorized, "invalid credentials")
			return
		default:
			log.Printf("SignIn error: %v", err)
			response.RespondWithError(w, http.StatusInternalServerError, "internal server error")
			return
		}
	}

	// Valid credentials -> create/renew the session
	err = h.sessionManager.RenewToken(ctx)
	if err != nil {
		log.Printf("Failed to create session token: %v", err)
		response.RespondWithError(w, http.StatusInternalServerError, "failed to create session")
		return
	}

	// Store only simple data in the session
	h.sessionManager.Put(ctx, "user_id", userID.String())

	// Return success
	responseJSON := map[string]interface{}{
		"message": "user successfully signed in",
		"data": map[string]string{
			"user_id": userID.String(),
		},
	}
	response.RespondWithJSON(w, http.StatusOK, responseJSON)
}

// SignOut logs the user out by destroying the session
func (h *Handler) SignOut(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.RespondWithError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	ctx := r.Context()

	// If we reach here, user is authenticated, so let's destroy the session
	err := h.sessionManager.Destroy(ctx)
	if err != nil {
		log.Printf("SignOut error: %v", err)
		response.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	responseJSON := map[string]string{"message": "successfully signed out"}
	response.RespondWithJSON(w, http.StatusOK, responseJSON)
}

// GetCurrentSession returns the current logged in user session data
func (h *Handler) GetCurrentSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.RespondWithError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	ctx := r.Context()
	userID := h.sessionManager.GetString(ctx, "user_id")

	if userID == "" {
		// Should never happen because of the auth check, but just in case
		response.RespondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	// Return success with current user_id
	responseJSON := map[string]interface{}{
		"message": "successfully get current session",
		"data": map[string]string{
			"user_id": userID,
		},
	}
	response.RespondWithJSON(w, http.StatusOK, responseJSON)
}
