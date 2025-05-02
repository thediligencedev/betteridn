# BetterIDN API Documentation

This directory contains the API documentation for the BetterIDN application, a social media/forum backend built with Go.

## API Overview

BetterIDN provides a RESTful API that enables clients to interact with the platform's features:

- **Authentication**: User registration, login, and session management
- **Posts**: Creating, reading, updating, and voting on posts

## Base URL

All API endpoints are prefixed with `/api/v1`.

## Authentication

The API uses cookie-based authentication. When a user signs in, a session cookie is set in the response. Subsequent requests should include this cookie to authenticate the user.

## API Endpoints

- [Authentication](./authentication.md): User registration, login, and session management
- [Posts](./posts.md): Post creation, retrieval, updates, and voting

## Error Handling

The API returns appropriate HTTP status codes along with JSON responses for errors:

```json
{
  "error": "Error message describing what went wrong"
}
```

Common error status codes:
- `400 Bad Request`: Invalid input or validation error
- `401 Unauthorized`: Authentication required
- `403 Forbidden`: Insufficient permissions
- `404 Not Found`: Resource not found
- `409 Conflict`: Resource conflict (e.g., duplicate email)
- `500 Internal Server Error`: Server-side error

## Content Types

All requests and responses use JSON format with the `application/json` content type, except for file uploads which use `multipart/form-data`.

## Swagger Documentation

The API is documented using OpenAPI/Swagger. The Swagger UI is available at `/swagger/` when the server is running.

## Rate Limiting

Some endpoints, particularly for email confirmations, have rate limiting to prevent abuse.