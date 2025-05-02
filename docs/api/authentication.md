# Authentication API

The Authentication API provides endpoints for user registration, login, logout, and session management.

## Endpoints

### Sign Up

Creates a new user account with email/password.

- **URL**: `/api/v1/auth/signup`
- **Method**: `POST`
- **Authentication**: No
- **Request Body**:
  ```json
  {
    "username": "johndoe",
    "email": "johndoe@example.com",
    "password": "securepassword"
  }
  ```
- **Response**:
  - **Success (200)**:
    ```json
    {
      "message": "successfully created user, please check your email to confirm"
    }
    ```
  - **Error (400)**: Bad Request (validation error)
  - **Error (409)**: Conflict (user already exists)
  - **Error (500)**: Internal Server Error

### Sign In

Authenticates a user with email/password and creates a session.

- **URL**: `/api/v1/auth/signin`
- **Method**: `POST`
- **Authentication**: No
- **Request Body**:
  ```json
  {
    "email": "johndoe@example.com",
    "password": "securepassword"
  }
  ```
- **Response**:
  - **Success (200)**:
    ```json
    {
      "message": "user successfully signed in",
      "data": {
        "user_id": "4fa85f64-5717-4562-b3fc-2c963f66afa6"
      }
    }
    ```
  - **Success but Email Not Confirmed (200)**:
    ```json
    {
      "message": "user successfully signed in",
      "data": {
        "user_id": "4fa85f64-5717-4562-b3fc-2c963f66afa6"
      },
      "warning": "Your email is not yet confirmed. Please check your inbox."
    }
    ```
  - **Error (400)**: Bad Request (validation error)
  - **Error (401)**: Unauthorized (invalid credentials)
  - **Error (500)**: Internal Server Error

### Sign Out

Ends a user's session.

- **URL**: `/api/v1/auth/signout`
- **Method**: `POST`
- **Authentication**: Required
- **Response**:
  - **Success (200)**:
    ```json
    {
      "message": "successfully signed out"
    }
    ```
  - **Error (401)**: Unauthorized (not logged in)
  - **Error (500)**: Internal Server Error

### Get Current Session

Returns information about the current user session.

- **URL**: `/api/v1/auth/session`
- **Method**: `GET`
- **Authentication**: Required
- **Response**:
  - **Success (200)**:
    ```json
    {
      "message": "successfully get current session",
      "data": {
        "user_id": "4fa85f64-5717-4562-b3fc-2c963f66afa6"
      }
    }
    ```
  - **Error (401)**: Unauthorized (session not found)

### Google OAuth Login

Initiates the Google OAuth flow.

- **URL**: `/api/v1/auth/google/login`
- **Method**: `GET`
- **Authentication**: No
- **Response**:
  - **Success (302)**: Redirects to Google consent page

### Google OAuth Callback

Handles the callback from Google OAuth.

- **URL**: `/api/v1/auth/google/callback`
- **Method**: `GET`
- **Authentication**: No
- **Query Parameters**:
  - `code`: Authorization code from Google
  - `state`: State parameter to prevent CSRF
- **Response**:
  - **Success (302)**: Redirects to `/home` on successful authentication
  - **Error (400)**: Bad Request (invalid parameters)
  - **Error (500)**: Internal Server Error

### Confirm Email

Confirms a user's email address using the token sent via email.

- **URL**: `/api/v1/auth/confirm-email`
- **Method**: `GET`
- **Authentication**: No
- **Query Parameters**:
  - `token`: Email confirmation token
- **Response**:
  - **Success (200)**:
    ```json
    {
      "message": "email confirmed successfully"
    }
    ```
  - **Error (400)**: Bad Request (invalid or expired token)

### Resend Confirmation

Resends the email confirmation token.

- **URL**: `/api/v1/auth/resend-confirmation`
- **Method**: `POST`
- **Authentication**: Required
- **Response**:
  - **Success (200)**:
    ```json
    {
      "message": "confirmation email resent. check your inbox."
    }
    ```
  - **Error (401)**: Unauthorized (not logged in)
  - **Error (429)**: Too Many Requests (rate limit exceeded)
  - **Error (500)**: Internal Server Error

## Authentication Flow

1. User registers via `/api/v1/auth/signup`
2. User receives confirmation email and confirms via `/api/v1/auth/confirm-email?token=xxx`
3. User signs in via `/api/v1/auth/signin`
4. The server sets a session cookie
5. User's subsequent requests include this cookie for authentication
6. User can sign out via `/api/v1/auth/signout`

Alternatively, users can authenticate via Google OAuth by visiting `/api/v1/auth/google/login`.