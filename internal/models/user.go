package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/pgtype"
)

type User struct {
	ID               uuid.UUID    `db:"id" json:"id"`
	Username         string       `db:"username" json:"username"`
	Email            string       `db:"email" json:"email"`
	Password         string       `db:"password" json:"-"`
	IsEmailConfirmed bool         `db:"is_email_confirmed" json:"is_email_confirmed"`
	Bio              string       `db:"bio" json:"bio,omitempty"`
	AvatarURL        string       `db:"avatar_url" json:"avatar_url,omitempty"`
	Preferences      pgtype.JSONB `db:"preferences" json:"preferences,omitempty"`
	LastSeenAt       *time.Time   `db:"last_seen_at" json:"last_seen_at,omitempty"`
	CreatedAt        time.Time    `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time    `db:"updated_at" json:"updated_at"`
}

// to avoid naming collisions in context
// and to store scs session inside golang context
type contextKey string

const UserContextKey contextKey = "user_id"
