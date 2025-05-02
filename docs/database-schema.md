# Database Schema

This document describes the database schema used by BetterIDN, a social media/forum application. The schema is designed for PostgreSQL database.

## Tables

### Users

Stores user account information and profile data.

```sql
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
```

| Column | Type | Description |
| ------ | ---- | ----------- |
| id | UUID | Primary key, auto-generated |
| username | TEXT | Unique username |
| email | TEXT | Unique email address |
| password | TEXT | Hashed password |
| is_email_confirmed | BOOLEAN | Email verification status |
| bio | TEXT | User's biography |
| avatar_url | TEXT | Profile picture URL |
| preferences | JSONB | User preferences (JSON) |
| last_seen_at | TIMESTAMPTZ | Last activity timestamp |
| created_at | TIMESTAMPTZ | Account creation timestamp |
| updated_at | TIMESTAMPTZ | Account update timestamp |

### Email Confirmations

Tracks email verification tokens.

```sql
CREATE TABLE email_confirmations (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    token TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    last_sent_at TIMESTAMPTZ DEFAULT NOW(),
    is_stale BOOLEAN DEFAULT FALSE
);
```

| Column | Type | Description |
| ------ | ---- | ----------- |
| user_id | UUID | Foreign key to users.id |
| token | TEXT | Verification token |
| expires_at | TIMESTAMPTZ | Token expiration time |
| created_at | TIMESTAMPTZ | Token creation timestamp |
| last_sent_at | TIMESTAMPTZ | Last email send timestamp |
| is_stale | BOOLEAN | Indicates if token is stale |

### Posts

Stores main content posts created by users.

```sql
CREATE TABLE posts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    content TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ
);
```

| Column | Type | Description |
| ------ | ---- | ----------- |
| id | UUID | Primary key, auto-generated |
| user_id | UUID | Foreign key to users.id |
| title | TEXT | Post title |
| content | TEXT | Post content |
| created_at | TIMESTAMPTZ | Post creation timestamp |
| updated_at | TIMESTAMPTZ | Post update timestamp |

### Comments

Stores user comments on posts.

```sql
CREATE TABLE comments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ
);
```

| Column | Type | Description |
| ------ | ---- | ----------- |
| id | UUID | Primary key, auto-generated |
| post_id | UUID | Foreign key to posts.id |
| user_id | UUID | Foreign key to users.id |
| content | TEXT | Comment content |
| created_at | TIMESTAMPTZ | Comment creation timestamp |
| updated_at | TIMESTAMPTZ | Comment update timestamp |

### Post Comments Metadata

Tracks metadata about comments on posts.

```sql
CREATE TABLE post_comments_metadata (
    post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    first_comment_id UUID REFERENCES comments(id),
    last_comment_id UUID REFERENCES comments(id),
    first_comment_at TIMESTAMPTZ,
    last_comment_at TIMESTAMPTZ,
    PRIMARY KEY (post_id)
);
```

| Column | Type | Description |
| ------ | ---- | ----------- |
| post_id | UUID | Foreign key to posts.id |
| first_comment_id | UUID | First comment's ID |
| last_comment_id | UUID | Most recent comment's ID |
| first_comment_at | TIMESTAMPTZ | First comment timestamp |
| last_comment_at | TIMESTAMPTZ | Most recent comment timestamp |

### Post Votes

Tracks upvotes and downvotes on posts.

```sql
CREATE TABLE post_votes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    vote_type INT CHECK (vote_type IN (-1, 1)),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (post_id, user_id)
);
```

| Column | Type | Description |
| ------ | ---- | ----------- |
| id | UUID | Primary key, auto-generated |
| post_id | UUID | Foreign key to posts.id |
| user_id | UUID | Foreign key to users.id |
| vote_type | INT | Vote type: 1 (upvote) or -1 (downvote) |
| created_at | TIMESTAMPTZ | Vote timestamp |

### Comment Votes

Tracks upvotes and downvotes on comments.

```sql
CREATE TABLE comment_votes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    comment_id UUID NOT NULL REFERENCES comments(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    vote_type INT CHECK (vote_type IN (-1, 1)),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (comment_id, user_id)
);
```

| Column | Type | Description |
| ------ | ---- | ----------- |
| id | UUID | Primary key, auto-generated |
| comment_id | UUID | Foreign key to comments.id |
| user_id | UUID | Foreign key to users.id |
| vote_type | INT | Vote type: 1 (upvote) or -1 (downvote) |
| created_at | TIMESTAMPTZ | Vote timestamp |

### Login Providers

Tracks third-party authentication providers.

```sql
CREATE TABLE login_providers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider TEXT NOT NULL,
    identifier TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (user_id, provider)
);
```

| Column | Type | Description |
| ------ | ---- | ----------- |
| id | UUID | Primary key, auto-generated |
| user_id | UUID | Foreign key to users.id |
| provider | TEXT | Provider name (e.g., 'google') |
| identifier | TEXT | User identifier from provider |
| created_at | TIMESTAMPTZ | Creation timestamp |

### Categories

Stores post categories.

```sql
CREATE TABLE categories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
```

| Column | Type | Description |
| ------ | ---- | ----------- |
| id | UUID | Primary key, auto-generated |
| name | TEXT | Category name |
| created_at | TIMESTAMPTZ | Creation timestamp |

### Post Categories

Many-to-many relationship between posts and categories.

```sql
CREATE TABLE post_categories (
    post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    category_id UUID NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
    PRIMARY KEY (post_id, category_id)
);
```

| Column | Type | Description |
| ------ | ---- | ----------- |
| post_id | UUID | Foreign key to posts.id |
| category_id | UUID | Foreign key to categories.id |

### Notifications

Tracks user notifications.

```sql
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

CREATE INDEX idx_notifications_user_id ON notifications(user_id);
CREATE INDEX idx_notifications_read_at ON notifications(read_at);
```

| Column | Type | Description |
| ------ | ---- | ----------- |
| id | UUID | Primary key, auto-generated |
| user_id | UUID | Recipient user ID |
| from_user_id | UUID | Sender user ID |
| type | TEXT | Notification type |
| subject_type | TEXT | Subject type |
| subject_id | UUID | Subject ID |
| data | JSONB | Additional data (JSON) |
| created_at | TIMESTAMPTZ | Creation timestamp |
| read_at | TIMESTAMPTZ | When notification was read |
| is_deleted | BOOLEAN | Soft delete flag |

### Sessions

Stores server-side sessions.

```sql
CREATE TABLE sessions (
    token TEXT PRIMARY KEY,
    data BYTEA NOT NULL,
    expiry TIMESTAMPTZ NOT NULL
);

CREATE INDEX sessions_expiry_idx ON sessions (expiry);
```

| Column | Type | Description |
| ------ | ---- | ----------- |
| token | TEXT | Session token |
| data | BYTEA | Session data |
| expiry | TIMESTAMPTZ | Expiration timestamp |

## Relationships

### Entity Relationship Diagram

```
users ----1:M----> posts
users ----1:M----> comments
users ----1:M----> post_votes
users ----1:M----> comment_votes
users ----1:1----> email_confirmations
users ----1:M----> login_providers
users ----1:M----> notifications (recipient)
users ----1:M----> notifications (sender)

posts ----1:M----> comments
posts ----1:M----> post_votes
posts ----1:1----> post_comments_metadata
posts ----M:M----> categories (via post_categories)

comments ----1:M----> comment_votes
```

## Indexes

- Primary keys on all tables
- Foreign key relationships for referential integrity
- Unique constraints on username, email, session token
- Index on notifications.user_id for quick lookup
- Index on notifications.read_at for filtering
- Index on sessions.expiry for cleanup

## Data Types

- UUIDs for all primary keys (except sessions)
- TEXT for string data (variable length)
- BOOLEAN for flags
- TIMESTAMPTZ for all timestamps (timezone aware)
- JSONB for structured data
- BYTEA for binary data

## Constraints

- Foreign key constraints with CASCADE deletions
- Check constraints on vote_type, notification_type
- Unique constraints to prevent duplicates