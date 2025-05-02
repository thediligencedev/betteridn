# BetterIDN Technical Architecture

This document describes the technical architecture of the BetterIDN application, a social media/forum backend built with Go.

## Overview

BetterIDN is a RESTful API-based application that follows a clean architecture pattern with:

1. Domain-driven design
2. Separation of concerns
3. Dependency injection
4. Service-based architecture

## Key Components

### Server Layer (`internal/server`)

- Entry point for the application
- HTTP server configuration
- Route registration
- Middleware configuration
- Static file serving
- OpenAPI documentation via Swagger UI

### Handler Layer (`internal/auth`, `internal/post`)

- HTTP request handling
- Request validation
- Response formatting
- Error handling
- Authentication checks

### Service Layer (`internal/auth`, `internal/post`)

- Business logic
- Domain operations
- Transaction management
- Error wrapping
- Integration with infrastructure

### Model Layer (`internal/models`)

- Domain models
- Data structures
- Type definitions

### Database Layer (`internal/db`)

- Database connection management
- Connection pooling
- Transaction support
- Migration support

### Worker Layer (`internal/worker`)

- Background job processing
- Email sending
- Asynchronous tasks

### Configuration (`internal/config`)

- Environment-based configuration
- Defaults and overrides
- Secrets management

## Authentication & Authorization

BetterIDN uses cookie-based authentication with server-side sessions:

1. Sessions are stored in PostgreSQL database
2. Session cookies are secure, HTTP-only
3. CSRF protection is implemented
4. Authentication middleware validates sessions
5. Google OAuth is supported for third-party login

## Data Flow

1. Client sends HTTP request to API endpoint
2. Server routes the request to appropriate handler
3. Middleware performs authentication and logging
4. Handler parses and validates the request
5. Handler delegates to service layer for business logic
6. Service interacts with database as needed
7. Response travels back through the layers
8. Handler formats and returns the response

## Middleware Stack

- Logging middleware: Records requests and responses
- Authentication middleware: Validates user sessions
- CORS middleware: Handles cross-origin requests
- OPTIONS responder: Handles preflight requests

## Database Schema

The application uses PostgreSQL with the following key tables:

- `users`: User accounts and profiles
- `email_confirmations`: Email verification tokens
- `posts`: User-created content
- `comments`: Responses to posts
- `post_votes`: Upvotes/downvotes on posts
- `comment_votes`: Upvotes/downvotes on comments
- `categories`: Topic categories
- `post_categories`: Many-to-many relationship
- `notifications`: User notifications
- `sessions`: Server-side session storage
- `login_providers`: OAuth provider details

## Security Considerations

- Password hashing with bcrypt
- Email verification flow
- Rate limiting on sensitive endpoints
- HTTPS support
- Session security (HTTP-only, secure flags)
- Input validation
- SQL injection prevention
- XSS protection
- CSRF protection

## Deployment Architecture

The application is designed to be deployed as:

1. Containerized application (Docker)
2. Database in separate container or managed service
3. Horizontally scalable API servers
4. Load balancer for distribution
5. Static files served through CDN

## Dependencies

Key external dependencies:

- **gorilla/mux**: HTTP routing
- **pgx**: PostgreSQL driver
- **scs**: Session management
- **zap**: Logging
- **golang-migrate**: Database migrations
- **validator**: Input validation

## Error Handling

The application uses a structured approach to error handling:

1. Domain-specific error types
2. Error wrapping with context
3. Consistent error responses
4. Logging at appropriate levels

## Development Workflow

1. Code organization by domain (auth, post, etc.)
2. Dependency injection for testability
3. Database migrations for schema changes
4. Configuration via environment variables
5. Local development with hot reloading