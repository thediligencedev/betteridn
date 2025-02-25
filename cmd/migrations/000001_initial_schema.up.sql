-- pkg/db/migrations/0001_initial_schema_up.sql

-- Enable UUID-OSSP extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Table: users
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username TEXT NOT NULL UNIQUE,
    email TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL,
    is_email_confirmed BOOLEAN DEFAULT FALSE,
    bio TEXT,
    avatar_url TEXT,
    preferences JSONB,
    last_seen_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Table: email_confirmations
CREATE TABLE email_confirmations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Table: posts
CREATE TABLE posts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    content TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    edited_at TIMESTAMPTZ
);

-- Table: comments
CREATE TABLE comments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    edited_at TIMESTAMPTZ
);

-- Table: post_comments_metadata
CREATE TABLE post_comments_metadata (
    post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    first_comment_id UUID REFERENCES comments(id),
    last_comment_id UUID REFERENCES comments(id),
    first_comment_at TIMESTAMPTZ,
    last_comment_at TIMESTAMPTZ,
    PRIMARY KEY (post_id)
);

-- Table: post_votes
CREATE TABLE post_votes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    vote_type INT CHECK (vote_type IN (-1, 1)),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (post_id, user_id)  -- Prevent multiple votes by the same user
);

-- Table: comment_votes
CREATE TABLE comment_votes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    comment_id UUID NOT NULL REFERENCES comments(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    vote_type INT CHECK (vote_type IN (-1, 1)),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (comment_id, user_id)  -- Prevent multiple votes by the same user
);

-- Table: login_providers
CREATE TABLE login_providers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider TEXT NOT NULL,
    identifier TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (user_id, provider)  -- Ensure no duplicate providers per user
);

-- Table: tags
CREATE TABLE tags (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Table: post_tags
CREATE TABLE post_tags (
    post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    tag_id UUID NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (post_id, tag_id)
);

-- Table: notifications
CREATE TABLE notifications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    from_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type TEXT NOT NULL CHECK (type IN ('post_like', 'comment_like', 'new_comment', 'mention', 'follow')),
    subject_type TEXT NOT NULL CHECK (subject_type IN ('post', 'comment', 'user')),
    subject_id UUID NOT NULL,
    data JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    read_at TIMESTAMPTZ,
    is_deleted BOOLEAN DEFAULT FALSE
);

-- Indexes for notifications
CREATE INDEX idx_notifications_user_id ON notifications(user_id);
CREATE INDEX idx_notifications_read_at ON notifications(read_at);


-- Table: sessions
CREATE TABLE sessions (
    token TEXT PRIMARY KEY,
    data BYTEA NOT NULL,
    expiry TIMESTAMPTZ NOT NULL
);

-- Indexes for sessions
CREATE INDEX sessions_expiry_idx ON sessions (expiry);


-- CREATE TABLE sessions (
--     id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
--     user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
--     token TEXT NOT NULL UNIQUE,
--     expires_at TIMESTAMPTZ NOT NULL,
--     created_at TIMESTAMPTZ DEFAULT NOW(),
--     updated_at TIMESTAMPTZ DEFAULT NOW()
-- );



-- Indexes for sessions
-- CREATE INDEX idx_sessions_id ON sessions(id);
-- CREATE INDEX idx_sessions_token ON sessions(token);


