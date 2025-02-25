package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/thediligencedev/betteridn/internal/worker"
	// any other imports you need
)

// ConfirmationService handles email confirmation logic.
type ConfirmationService struct {
	pool        *pgxpool.Pool
	emailWorker *worker.EmailWorker
}

// NewConfirmationService creates a new ConfirmationService using a *pgxpool.Pool.
func NewConfirmationService(pool *pgxpool.Pool, emailWorker *worker.EmailWorker) *ConfirmationService {
	return &ConfirmationService{
		pool:        pool,
		emailWorker: emailWorker,
	}
}

// GenerateAndSendConfirmation generates a token, upserts into email_confirmations, and sends
// a confirmation email. If an existing non-stale token is present, it checks the 5-minute rate limit.
func (cs *ConfirmationService) GenerateAndSendConfirmation(ctx context.Context, userID uuid.UUID, userEmail string) error {
	var existingToken string
	var lastSentAt time.Time
	var isStale bool

	// 1. See if there's an existing record for this user:
	row := cs.pool.QueryRow(ctx, `
		SELECT token, last_sent_at, is_stale
		FROM email_confirmations
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`, userID)
	err := row.Scan(&existingToken, &lastSentAt, &isStale)

	if err == nil {
		// We have an existing record
		if !isStale {
			// check rate limit
			if time.Since(lastSentAt) < 5*time.Minute {
				// less than 5 minutes
				return fmt.Errorf("please wait a few minutes before requesting another confirmation email")
			}
		}
		// If isStale=true or older than 5 minutes, we proceed with a new token.
	} else {
		// If err != pgx.ErrNoRows but some other error, bail out:
		if err.Error() != "no rows in result set" {
			return fmt.Errorf("error checking existing token: %w", err)
		}
		// If pgx.ErrNoRows, we continue.
	}

	// 2. Generate a new random token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return fmt.Errorf("failed to generate token: %w", err)
	}
	token := base64.URLEncoding.EncodeToString(tokenBytes)

	// 3. Insert/Update email_confirmations for this user
	expiresAt := time.Now().Add(24 * time.Hour)
	_, err = cs.pool.Exec(ctx, `
		INSERT INTO email_confirmations (user_id, token, expires_at, is_stale, last_sent_at)
		VALUES ($1, $2, $3, false, now())
		ON CONFLICT (user_id) DO UPDATE
		SET token = EXCLUDED.token,
		    expires_at = EXCLUDED.expires_at,
		    is_stale = false,
		    last_sent_at = now()
	`, userID, token, expiresAt)
	if err != nil {
		return fmt.Errorf("failed to insert/update email confirmation: %w", err)
	}

	// 4. Send the email via the worker
	subject := "Confirm Your Email Address"
	confirmLink := fmt.Sprintf("http://localhost:8080/api/v1/auth/confirm-email?token=%s", token)
	bodyHTML := strings.ReplaceAll(`
<!DOCTYPE html>
<html>
  <body>
    <h2>Confirm Your Email</h2>
    <p>Click the link below to confirm your email:</p>
    <a href="$LINK">Confirm Email</a>
    <p>This link will expire in 24 hours.</p>
  </body>
</html>
`, "$LINK", confirmLink)

	cs.emailWorker.Enqueue(worker.EmailJob{
		To:       userEmail,
		Subject:  subject,
		BodyHTML: bodyHTML,
	})

	return nil
}

// ConfirmEmailByToken sets user.is_email_confirmed = true if token is valid and not expired or stale.
func (cs *ConfirmationService) ConfirmEmailByToken(ctx context.Context, token string) error {
	// 1. Fetch email_confirmations record by token
	var userID uuid.UUID
	var expiresAt time.Time
	var isStale bool

	err := cs.pool.QueryRow(ctx, `
		SELECT user_id, expires_at, is_stale
		FROM email_confirmations
		WHERE token = $1
	`, token).Scan(&userID, &expiresAt, &isStale)
	if err != nil {
		return fmt.Errorf("invalid or unknown token")
	}
	if isStale {
		return fmt.Errorf("this token is stale or already used")
	}
	if time.Now().After(expiresAt) {
		// mark stale
		_, _ = cs.pool.Exec(ctx, `
			UPDATE email_confirmations
			SET is_stale = true
			WHERE token = $1
		`, token)
		return fmt.Errorf("this token has expired, please request a new confirmation email")
	}

	// 2. Mark user is_email_confirmed = true
	_, err = cs.pool.Exec(ctx, `
		UPDATE users
		SET is_email_confirmed = true
		WHERE id = $1
	`, userID)
	if err != nil {
		return fmt.Errorf("failed to confirm user: %w", err)
	}

	// 3. Mark the token as stale
	_, err = cs.pool.Exec(ctx, `
		UPDATE email_confirmations
		SET is_stale = true
		WHERE token = $1
	`, token)
	if err != nil {
		return fmt.Errorf("failed to mark token stale: %w", err)
	}

	return nil
}
