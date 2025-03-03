package models

import (
	"time"

	"github.com/google/uuid"
)

type Post struct {
	ID         uuid.UUID  `db:"id" json:"id"`
	Title      string     `db:"title" json:"title"`
	Content    string     `db:"content" json:"content"`
	CreatedAt  time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time  `db:"updated_at" json:"updated_at"`
	Categories []string   `json:"categories,omitempty"`
	User       *UserBasic `json:"user,omitempty"`
	VoteCount  *VoteCount `json:"vote_count,omitempty"`
}

type UserBasic struct {
	Username string `json:"username"`
}

type VoteCount struct {
	Upvotes   int `json:"upvotes"`
	Downvotes int `json:"downvotes"`
}
