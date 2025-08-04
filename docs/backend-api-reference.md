# BetterIDN Backend API Reference

This document provides a comprehensive overview of the BetterIDN backend API for frontend developers. It covers authentication, endpoints, data models, and integration guidelines.

## Overview

The BetterIDN backend is a Go-based API server using PostgreSQL as the database. It provides authentication, user management, and post/comment functionality with voting capabilities.

### Base URL
```
http://localhost:8080
```

### Tech Stack
- **Language**: Go 1.24
- **Database**: PostgreSQL with UUID support
- **Authentication**: Session-based (SCS library) with cookie storage
- **Password Hashing**: bcrypt
- **OAuth**: Google OAuth 2.0

## Authentication

The API uses session-based authentication with secure HTTP-only cookies. Sessions are stored in PostgreSQL and managed by the SCS library.

### Session Management
- **Cookie Name**: `session`
- **Session Expiry**: Configurable via `SESSION_EXPIRY` env variable
- **Credentials**: Must include `credentials: 'include'` in fetch requests

### Authentication Flow

#### 1. Email-based Registration
Users register with username, email, and password. Email confirmation is required before full access.

#### 2. Google OAuth
Users can sign in with Google. New Google users are automatically created with confirmed email status.

#### 3. Session Persistence
After successful authentication, a session cookie is set. All protected endpoints require this session.

## API Endpoints

### Authentication Endpoints

#### Sign Up
```http
POST /api/v1/auth/signup
Content-Type: application/json

{
  "username": "string",
  "email": "string",
  "password": "string (min 6 chars)"
}

Response:
{
  "message": "successfully created user, please check your email to confirm"
}
```

#### Sign In
```http
POST /api/v1/auth/signin
Content-Type: application/json

{
  "email": "string",
  "password": "string"
}

Response:
{
  "message": "user successfully signed in",
  "data": {
    "user_id": "uuid"
  },
  "warning": "Your email is not yet confirmed. Please check your inbox." // if email not confirmed
}
```

#### Sign Out
```http
POST /api/v1/auth/signout
(Requires authentication)

Response:
{
  "message": "successfully signed out"
}
```

#### Get Current Session
```http
GET /api/v1/auth/session
(Requires authentication)

Response:
{
  "message": "successfully get current session",
  "data": {
    "user_id": "uuid"
  }
}
```

#### Google OAuth Login
```http
GET /api/v1/auth/google/login

Redirects to Google OAuth consent page
```

#### Google OAuth Callback
```http
GET /api/v1/auth/google/callback?code=...&state=...

Handles OAuth callback and redirects to /home on success
```

#### Confirm Email
```http
GET /api/v1/auth/confirm-email?token=...

Response:
{
  "message": "email confirmed successfully"
}
```

#### Resend Confirmation Email
```http
POST /api/v1/auth/resend-confirmation
(Requires authentication)

Response:
{
  "message": "confirmation email resent. check your inbox."
}
```

### Post Endpoints

#### Create Post
```http
POST /api/v1/posts
Content-Type: application/json
(Requires authentication)

{
  "title": "string",
  "content": "string",
  "categories": ["string"] // must have at least 1 category
}

Response:
{
  "message": "post created successfully",
  "data": {
    "id": "uuid"
  }
}
```

#### Get Posts (Paginated)
```http
GET /api/v1/posts?page=1&limit=20
(Optional authentication)

Response:
{
  "message": "posts retrieved successfully",
  "data": [
    {
      "id": "uuid",
      "title": "string",
      "content": "string",
      "created_at": "timestamp",
      "updated_at": "timestamp",
      "categories": ["string"],
      "user": {
        "username": "string"
      },
      "vote_count": {
        "upvotes": number,
        "downvotes": number
      }
    }
  ]
}
```

#### Get Post by ID
```http
GET /api/v1/posts/{postId}
(Optional authentication)

Response:
{
  "message": "post retrieved successfully",
  "data": {
    "id": "uuid",
    "title": "string",
    "content": "string",
    "created_at": "timestamp",
    "updated_at": "timestamp",
    "categories": ["string"],
    "user": {
      "username": "string"
    },
    "vote_count": {
      "upvotes": number,
      "downvotes": number
    }
  }
}
```

#### Update Post
```http
PUT /api/v1/posts/{postId}
Content-Type: application/json
(Requires authentication - must be post owner)

{
  "title": "string",
  "content": "string",
  "categories": ["string"]
}

Response:
{
  "message": "post updated successfully",
  "data": {
    "id": "uuid"
  }
}
```

#### Vote on Post
```http
POST /api/v1/posts/{postId}/vote
Content-Type: application/json
(Requires authentication)

{
  "vote_type": 1 // 1 for upvote, -1 for downvote
}

Response:
{
  "message": "vote recorded successfully", // or "successfully removed vote"
  "data": {
    "id": "uuid",
    "vote_count": {
      "upvotes": number,
      "downvotes": number
    }
  }
}
```

Note: Voting again with the same vote_type removes the vote. Voting with a different vote_type changes the vote.

## Data Models

### User
```typescript
interface User {
  id: string;           // UUID
  username: string;
  email: string;
  is_email_confirmed: boolean;
  bio?: string;
  avatar_url?: string;
  preferences?: object; // JSONB
  last_seen_at?: string; // timestamp
  created_at: string;   // timestamp
  updated_at: string;   // timestamp
}
```

### Post
```typescript
interface Post {
  id: string;           // UUID
  title: string;
  content: string;
  created_at: string;   // timestamp
  updated_at: string;   // timestamp
  categories: string[];
  user: {
    username: string;
  };
  vote_count: {
    upvotes: number;
    downvotes: number;
  };
}
```

### Error Response
```typescript
interface ErrorResponse {
  error: string;
}
```

## CORS Configuration

The backend is configured to allow CORS from:
- Origin: `http://localhost:5500`
- Credentials: `true`
- Methods: `GET, POST, PUT, OPTIONS`
- Headers: `Content-Type, Authorization, hx-current-url, hx-request, hx-target, hx-trigger, Accept, Content-Length, Accept-Encoding, Accept-Language, Credentials`

## Frontend Integration Guidelines

### 1. Making Authenticated Requests

Always include credentials in fetch requests:

```javascript
fetch('http://localhost:8080/api/v1/posts', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
  },
  credentials: 'include', // Important for session cookies
  body: JSON.stringify({
    title: 'My Post',
    content: 'Post content',
    categories: ['general']
  })
});
```

### 2. Handling Session State

Check if user is authenticated:

```javascript
async function checkAuth() {
  try {
    const response = await fetch('http://localhost:8080/api/v1/auth/session', {
      credentials: 'include'
    });
    
    if (response.ok) {
      const data = await response.json();
      return { isAuthenticated: true, userId: data.data.user_id };
    }
    return { isAuthenticated: false };
  } catch (error) {
    return { isAuthenticated: false };
  }
}
```

### 3. Error Handling

All errors return a consistent format:

```javascript
async function handleApiCall(url, options = {}) {
  try {
    const response = await fetch(url, {
      ...options,
      credentials: 'include'
    });
    
    const data = await response.json();
    
    if (!response.ok) {
      throw new Error(data.error || 'API call failed');
    }
    
    return data;
  } catch (error) {
    console.error('API Error:', error);
    throw error;
  }
}
```

### 4. Pagination

When fetching posts, use query parameters:

```javascript
const page = 1;
const limit = 20;
const response = await fetch(`http://localhost:8080/api/v1/posts?page=${page}&limit=${limit}`, {
  credentials: 'include'
});
```

### 5. Real-time Features

Currently, the API doesn't support WebSockets. For real-time updates, consider:
- Polling for new posts/comments
- Refreshing vote counts after user actions
- Checking for new notifications periodically

## Important Notes

1. **Email Confirmation**: Users registered via email must confirm their email before certain features are available. Google OAuth users skip this requirement.

2. **Vote Behavior**: 
   - Users can upvote (+1) or downvote (-1) posts
   - Voting again with the same type removes the vote
   - Changing vote type updates the existing vote

3. **Categories**: Posts must have at least one category. Categories must exist in the database beforehand.

4. **Authentication State**: The backend uses HTTP-only cookies for sessions. Frontend cannot access the session token directly but must include credentials in all requests.

5. **User Context**: For protected endpoints, the user ID is automatically extracted from the session. No need to send user ID in request body.

## Environment Variables

The backend requires these environment variables:

```env
# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=your_user
DB_PASSWORD=your_password
DB_NAME=betteridn
DB_SSLMODE=disable

# Server
SERVER_PORT=8080
SERVER_ENV=development

# Session
SESSION_SECRET=your-secret-key
SESSION_EXPIRY=24h

# Google OAuth
GOOGLE_CLIENT_ID=your-client-id
GOOGLE_CLIENT_SECRET=your-client-secret
GOOGLE_REDIRECT_URL=http://localhost:8080/api/v1/auth/google/callback

# SMTP (for email confirmation)
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_FROM=noreply@example.com
SMTP_USER=smtp_user
SMTP_PASS=smtp_password
```

## Future Considerations

Based on the database schema, these features are planned but not yet implemented:

1. **Comments System**: Tables exist for comments with nested structure
2. **Notifications**: Table exists for user notifications (post likes, comments, mentions, follows)
3. **User Following**: Notification types suggest a follow system
4. **Comment Voting**: Similar to post voting but for comments

Frontend developers should be prepared for these features in future API updates.