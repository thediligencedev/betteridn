package auth

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/thediligencedev/betteridn/internal/models"
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
}

// NewAuthService creates a new AuthService.
func NewAuthService(pool *pgxpool.Pool) *AuthService {
	return &AuthService{pool: pool}
}

// SignUp creates a new user in the database and
// also creates a login_providers row with provider='email'.
func (s *AuthService) SignUp(ctx context.Context, username, email, plainPassword string) error {
	// 1. Check if user (by email or username) already exists
	var exists bool
	checkQuery := `
		SELECT EXISTS (
			SELECT 1 
			FROM users 
			WHERE LOWER(email)=$1 OR LOWER(username)=$2
		)
	`
	err := s.pool.QueryRow(ctx, checkQuery, strings.ToLower(email), strings.ToLower(username)).Scan(&exists)
	if err != nil {
		return err
	}
	if exists {
		return ErrUserAlreadyExists
	}

	// 2. Hash password
	hashed, err := password.HashPassword(plainPassword)
	if err != nil {
		return err
	}

	// 3. Insert user, returning the new user id
	insertUserQuery := `
		INSERT INTO users (username, email, password) 
		VALUES ($1, $2, $3)
		RETURNING id
	`
	var userID uuid.UUID
	err = s.pool.QueryRow(ctx, insertUserQuery, username, email, hashed).Scan(&userID)
	if err != nil {
		return ErrCreateUser
	}

	// 4. Insert into login_providers with provider='email' and identifier as the email
	insertLPQuery := `
		INSERT INTO login_providers (user_id, provider, identifier)
		VALUES ($1, 'email', $2)
	`
	_, err = s.pool.Exec(ctx, insertLPQuery, userID, email)
	if err != nil {
		return ErrCreateUser
	}

	return nil
}

// SignIn checks user credentials and returns user_id if success.
func (s *AuthService) SignIn(ctx context.Context, email, plainPassword string) (uuid.UUID, error) {
	var user models.User
	query := `
		SELECT id, password
		FROM users
		WHERE LOWER(email)=$1
	`
	err := s.pool.QueryRow(ctx, query, strings.ToLower(email)).Scan(&user.ID, &user.Password)
	if err != nil {
		return uuid.Nil, ErrInvalidCredentials
	}

	// Compare password
	if err := password.CheckPassword(user.Password, plainPassword); err != nil {
		return uuid.Nil, ErrInvalidCredentials
	}
	return user.ID, nil
}
