package auth

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/alexedwards/scs/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/thediligencedev/betteridn/pkg/response"
	"github.com/thediligencedev/betteridn/pkg/validator"
)

// Handler holds dependencies for auth handlers.
type Handler struct {
	service        *AuthService
	sessionManager *scs.SessionManager
	confService    *ConfirmationService
}

// NewHandler modifies to accept ConfirmationService as well
func NewHandler(
	pool *pgxpool.Pool,
	sessionManager *scs.SessionManager,
	cs *ConfirmationService,
) *Handler {
	return &Handler{
		service:        NewAuthService(pool, cs),
		sessionManager: sessionManager,
		confService:    cs,
	}
}

type SignUpRequest struct {
	Username string `json:"username" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

type SignInRequest struct {
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// SignUp -> sign up user, send confirmation
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

	ctx := r.Context()
	err := h.service.SignUp(ctx, req.Username, req.Email, req.Password)
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
			response.RespondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	// Success
	responseJSON := map[string]string{"message": "successfully created user, please check your email to confirm"}
	response.RespondWithJSON(w, http.StatusOK, responseJSON)
}

// SignIn -> user signs in
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
	currentUserID := h.sessionManager.GetString(ctx, "user_id")
	if currentUserID != "" {
		// Already logged in
		err := h.sessionManager.RenewToken(ctx)
		if err != nil {
			log.Printf("Failed to renew session token: %v", err)
			response.RespondWithError(w, http.StatusInternalServerError, "internal server error")
			return
		}
		responseJSON := map[string]interface{}{
			"message": "user already signed in",
			"data": map[string]string{
				"user_id": currentUserID,
			},
		}
		response.RespondWithJSON(w, http.StatusOK, responseJSON)
		return
	}

	// Attempt sign in
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

	//  Create session
	err = h.sessionManager.RenewToken(ctx)
	if err != nil {
		log.Printf("Failed to create session token: %v", err)
		response.RespondWithError(w, http.StatusInternalServerError, "failed to create session")
		return
	}
	h.sessionManager.Put(ctx, "user_id", userID.String())

	//  Check if user is confirmed
	var isConfirmed bool
	qErr := h.service.pool.QueryRow(ctx,
		`SELECT is_email_confirmed FROM users WHERE id = $1`, userID).Scan(&isConfirmed)
	if qErr != nil {
		log.Printf("error checking is_email_confirmed: %v", qErr)
	}

	//  Return success + a warning if not confirmed
	responseJSON := map[string]interface{}{
		"message": "user successfully signed in",
		"data": map[string]string{
			"user_id": userID.String(),
		},
	}
	if !isConfirmed {
		responseJSON["warning"] = "Your email is not yet confirmed. Please check your inbox."
	}

	response.RespondWithJSON(w, http.StatusOK, responseJSON)
}

// SignOut -> destroy session
func (h *Handler) SignOut(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.RespondWithError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	ctx := r.Context()
	if err := h.sessionManager.Destroy(ctx); err != nil {
		log.Printf("SignOut error: %v", err)
		response.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	responseJSON := map[string]string{"message": "successfully signed out"}
	response.RespondWithJSON(w, http.StatusOK, responseJSON)
}

// GetCurrentSession -> return current user session data
func (h *Handler) GetCurrentSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.RespondWithError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	ctx := r.Context()
	userID := h.sessionManager.GetString(ctx, "user_id")
	if userID == "" {
		response.RespondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	responseJSON := map[string]interface{}{
		"message": "successfully get current session",
		"data":    map[string]string{"user_id": userID},
	}
	response.RespondWithJSON(w, http.StatusOK, responseJSON)
}

// ConfirmEmail -> GET /api/v1/auth/confirm-email?token=xxx
func (h *Handler) ConfirmEmail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.RespondWithError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	token := r.URL.Query().Get("token")
	if token == "" {
		response.RespondWithError(w, http.StatusBadRequest, "missing token")
		return
	}

	ctx := r.Context()
	err := h.confService.ConfirmEmailByToken(ctx, token)
	if err != nil {
		// We might show the error or redirect to an error page
		response.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	// success
	responseJSON := map[string]string{"message": "email confirmed successfully"}
	response.RespondWithJSON(w, http.StatusOK, responseJSON)
}

// ResendConfirmation -> POST /api/v1/auth/resend-confirmation
func (h *Handler) ResendConfirmation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.RespondWithError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	ctx := r.Context()

	// Check if user is logged in (or pass email in body). Up to your design.
	userID := h.sessionManager.GetString(ctx, "user_id")
	if userID == "" {
		response.RespondWithError(w, http.StatusUnauthorized, "not logged in")
		return
	}
	uID, err := uuid.Parse(userID)
	if err != nil {
		response.RespondWithError(w, http.StatusBadRequest, "invalid user_id session")
		return
	}

	// fetch user's email from DB
	var emailStr string
	err = h.service.pool.QueryRow(ctx, `SELECT email FROM users WHERE id = $1`, uID).Scan(&emailStr)
	if err != nil {
		response.RespondWithError(w, http.StatusNotFound, "user not found")
		return
	}

	// call GenerateAndSendConfirmation
	err = h.confService.GenerateAndSendConfirmation(ctx, uID, emailStr)
	if err != nil {
		response.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	responseJSON := map[string]string{"message": "confirmation email resent. check your inbox."}
	response.RespondWithJSON(w, http.StatusOK, responseJSON)
}
