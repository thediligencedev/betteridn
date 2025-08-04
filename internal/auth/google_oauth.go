package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/alexedwards/scs/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/thediligencedev/betteridn/internal/config"
	"github.com/thediligencedev/betteridn/pkg/response"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var googleOAuthConfig *oauth2.Config

func InitGoogleOAuth(cfg *config.Config) {
	googleOAuthConfig = &oauth2.Config{
		ClientID:     cfg.GoogleClientID,
		ClientSecret: cfg.GoogleClientSecret,
		Endpoint:     google.Endpoint,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		RedirectURL: cfg.GoogleOAuthRedirectURL,
	}
}

// GoogleHandler no longer needs a ConfirmationService since we do not confirm OAuth emails
type GoogleHandler struct {
	pool           *pgxpool.Pool
	sessionManager *scs.SessionManager
	cfg            *config.Config
}

// Revert to the old constructor without confirmationService
func NewGoogleHandler(pool *pgxpool.Pool, sessionManager *scs.SessionManager, cfg *config.Config) *GoogleHandler {
	return &GoogleHandler{
		pool:           pool,
		sessionManager: sessionManager,
		cfg:            cfg,
	}
}

// GoogleLogin -> redirect to Google
func (gh *GoogleHandler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	if googleOAuthConfig == nil {
		response.RespondWithError(w, http.StatusInternalServerError, "Google OAuth not initialized")
		return
	}

	state, err := generateRandomState(16)
	if err != nil {
		log.Printf("Error generating random state: %v", err)
		response.RespondWithError(w, http.StatusInternalServerError, "failed to initiate google oauth")
		return
	}
	gh.sessionManager.Put(r.Context(), "oauth_state", state)

	url := googleOAuthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// GoogleCallback is the redirect URI set in your Google OAuth config
func (gh *GoogleHandler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	sessionState := gh.sessionManager.GetString(ctx, "oauth_state")
	queryState := r.URL.Query().Get("state")
	if sessionState == "" || sessionState != queryState {
		response.RespondWithError(w, http.StatusBadRequest, "invalid oauth state")
		return
	}
	gh.sessionManager.Remove(ctx, "oauth_state")

	code := r.URL.Query().Get("code")
	if code == "" {
		response.RespondWithError(w, http.StatusBadRequest, "missing code in callback")
		return
	}

	token, err := googleOAuthConfig.Exchange(ctx, code)
	if err != nil {
		log.Printf("Google token exchange error: %v", err)
		response.RespondWithError(w, http.StatusInternalServerError, "failed to exchange token")
		return
	}

	userInfo, err := fetchGoogleUserInfo(ctx, token)
	if err != nil {
		log.Printf("Failed to fetch google user info: %v", err)
		response.RespondWithError(w, http.StatusInternalServerError, "failed to fetch google userinfo")
		return
	}

	// find or create local user
	userID, err := gh.findOrCreateGoogleUser(ctx, userInfo)
	if err != nil {
		log.Printf("findOrCreateGoogleUser error: %v", err)
		response.RespondWithError(w, http.StatusInternalServerError, "failed to create or find user")
		return
	}

	// Renew session token
	if err := gh.sessionManager.RenewToken(ctx); err != nil {
		log.Printf("Failed to create session token: %v", err)
		response.RespondWithError(w, http.StatusInternalServerError, "failed to create session")
		return
	}

	gh.sessionManager.Put(ctx, "user_id", userID.String())

	// store refresh token if it exists
	if token.RefreshToken != "" {
		gh.sessionManager.Put(ctx, "google_refresh_token", token.RefreshToken)
	}

	// Redirect to frontend URL or fallback to /home
	redirectURL := gh.cfg.FrontendURL
	if redirectURL == "" {
		redirectURL = "/home"
	}
	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}

// fetch userinfo
func fetchGoogleUserInfo(ctx context.Context, token *oauth2.Token) (*GoogleUserInfo, error) {
	client := googleOAuthConfig.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, fmt.Errorf("failed to get userinfo: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("userinfo response status: %s", resp.Status)
	}

	var userInfo GoogleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("failed to decode userinfo: %w", err)
	}
	return &userInfo, nil
}

// minimal info from Google
type GoogleUserInfo struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

// For new google user -> is_email_confirmed = true
func (gh *GoogleHandler) findOrCreateGoogleUser(ctx context.Context, gu *GoogleUserInfo) (uuid.UUID, error) {
	// Check if we already have a user with this google login
	var existingUserID uuid.UUID
	err := gh.pool.QueryRow(ctx, `
        SELECT u.id
        FROM users u
        INNER JOIN login_providers lp ON lp.user_id = u.id
        WHERE lp.provider = 'google'
          AND lp.identifier = $1
    `, gu.Email).Scan(&existingUserID)

	if err == nil && existingUserID != uuid.Nil {
		// Found existing google user
		return existingUserID, nil
	}

	// If not found, check if there's a user with the same email from normal signup
	var emailUserID uuid.UUID
	err = gh.pool.QueryRow(ctx,
		`SELECT id FROM users WHERE lower(email)=lower($1)`,
		gu.Email).Scan(&emailUserID)
	if err == nil && emailUserID != uuid.Nil {
		// We have a user with that email but no google login_providers row
		err = gh.createLoginProvider(ctx, emailUserID, "google", gu.Email)
		if err != nil {
			return uuid.Nil, err
		}
		return emailUserID, nil
	}

	// otherwise, create a new user. For google, we do is_email_confirmed=true
	username := generateGoogleUsername(gu.Name, gu.Email)
	var newUserID uuid.UUID
	insertQ := `
        INSERT INTO users (username, email, password, is_email_confirmed)
        VALUES ($1, $2, $3, true)
        RETURNING id
    `
	err = gh.pool.QueryRow(ctx, insertQ, username, gu.Email, "oauth_no_password").Scan(&newUserID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to create new google user: %w", err)
	}

	// insert into login_providers
	err = gh.createLoginProvider(ctx, newUserID, "google", gu.Email)
	if err != nil {
		return uuid.Nil, err
	}

	return newUserID, nil
}

func (gh *GoogleHandler) createLoginProvider(ctx context.Context, userID uuid.UUID, provider, identifier string) error {
	insert := `
        INSERT INTO login_providers (user_id, provider, identifier)
        VALUES ($1, $2, $3)
        ON CONFLICT (user_id, provider) DO NOTHING
    `
	_, err := gh.pool.Exec(ctx, insert, userID, provider, identifier)
	return err
}

func generateGoogleUsername(name, email string) string {
	base := strings.ToLower(strings.ReplaceAll(name, " ", ""))
	if base == "" {
		parts := strings.Split(email, "@")
		base = parts[0]
	}
	base = sanitizeUsername(base)
	randDigits := randomDigits(4)
	return fmt.Sprintf("%s%s", base, randDigits)
}

func randomDigits(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	for i := 0; i < n; i++ {
		b[i] = '0' + (b[i] % 10)
	}
	return string(b)
}

func sanitizeUsername(s string) string {
	var sb strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			sb.WriteRune(r)
		}
	}
	return sb.String()
}

func generateRandomState(length int) (string, error) {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
