package auth

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/thediligencedev/betteridn/internal/models"
	"github.com/thediligencedev/betteridn/pkg/email"
	"github.com/thediligencedev/betteridn/pkg/password"
)

var (
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrCreateUser         = errors.New("failed to create user")
)

// AuthService defines methods for user authentication.
type AuthService struct {
	pool *pgxpool.Pool
	cs   *ConfirmationService // for email confirmation
}

func NewAuthService(pool *pgxpool.Pool, cs *ConfirmationService) *AuthService {
	return &AuthService{pool: pool, cs: cs}
}

// TODO: for signup and signin don't forget to lower the email and username

// SignUp (email-based). Now with domain checks + sending email confirm.
func (s *AuthService) SignUp(ctx context.Context, username, emailStr, plainPassword string) error {
	// 1. Basic check if user exists
	var exists bool
	checkQuery := `
        SELECT EXISTS (
            SELECT 1
            FROM users
            WHERE LOWER(email)=$1 OR LOWER(username)=$2
        )
    `
	err := s.pool.QueryRow(ctx, checkQuery, strings.ToLower(emailStr), strings.ToLower(username)).Scan(&exists)
	if err != nil {
		return err
	}
	if exists {
		return ErrUserAlreadyExists
	}

	// 2. Verify domain has MX/SPF/DMARC
	domain, err := email.ExtractDomain(emailStr)
	if err != nil {
		return err
	}
	if domainErr := email.IsDomainValid(domain); domainErr != nil {
		return domainErr
	}

	// 3. Hash password
	hashed, err := password.HashPassword(plainPassword)
	if err != nil {
		return err
	}

	// 4. Insert user: is_email_confirmed = false
	var userID uuid.UUID
	insertUserQuery := `
        INSERT INTO users (username, email, password, is_email_confirmed)
        VALUES ($1, $2, $3, false)
        RETURNING id
    `
	err = s.pool.QueryRow(ctx, insertUserQuery, username, emailStr, hashed).Scan(&userID)
	if err != nil {
		return ErrCreateUser
	}

	// 5. Also insert into login_providers with provider='email' and identifier as the email
	insertLPQuery := `
        INSERT INTO login_providers (user_id, provider, identifier)
        VALUES ($1, 'email', $2)
    `
	_, err = s.pool.Exec(ctx, insertLPQuery, userID, emailStr)
	if err != nil {
		return ErrCreateUser
	}

	// 6. Generate and send the email confirmation
	if err := s.cs.GenerateAndSendConfirmation(ctx, userID, emailStr); err != nil {
		// In production, might log the error but let the user proceed
		return err
	}

	return nil
}

// SignIn checks user credentials and returns user_id if success.
func (s *AuthService) SignIn(ctx context.Context, emailStr, plainPassword string) (uuid.UUID, error) {
	var user models.User
	query := `
        SELECT id, password
        FROM users
        WHERE LOWER(email)=$1
    `
	err := s.pool.QueryRow(ctx, query, strings.ToLower(emailStr)).Scan(&user.ID, &user.Password)
	if err != nil {
		return uuid.Nil, ErrInvalidCredentials
	}

	// Compare password
	if err := password.CheckPassword(user.Password, plainPassword); err != nil {
		return uuid.Nil, ErrInvalidCredentials
	}
	return user.ID, nil
}
