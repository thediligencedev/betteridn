# Posts API

The Posts API provides endpoints for creating, reading, updating, and voting on posts.

## Endpoints

### Create Post

Creates a new post.

- **URL**: `/api/v1/posts`
- **Method**: `POST`
- **Authentication**: Required
- **Request Body**:
  ```json
  {
    "title": "My First Post",
    "content": "This is the content of my first post.",
    "categories": ["technology", "golang"]
  }
  ```
- **Response**:
  - **Success (201)**:
    ```json
    {
      "message": "post created successfully",
      "data": {
        "id": "4fa85f64-5717-4562-b3fc-2c963f66afa6"
      }
    }
    ```
  - **Error (400)**: Bad Request (validation error or invalid request body)
  - **Error (401)**: Unauthorized (user not logged in)
  - **Error (500)**: Internal Server Error

### Get Posts

Retrieves a paginated list of posts.

- **URL**: `/api/v1/posts`
- **Method**: `GET`
- **Authentication**: Optional
- **Query Parameters**:
  - `page`: Page number (default: 1)
  - `limit`: Number of posts per page (default: 20, max: 100)
- **Response**:
  - **Success (200)**:
    ```json
    {
      "message": "posts retrieved successfully",
      "data": [
        {
          "id": "4fa85f64-5717-4562-b3fc-2c963f66afa6",
          "title": "My First Post",
          "content": "This is the content of my first post.",
          "created_at": "2023-04-01T12:00:00Z",
          "updated_at": "2023-04-01T12:00:00Z",
          "categories": ["technology", "golang"],
          "user": {
            "username": "johndoe"
          },
          "vote_count": {
            "upvotes": 10,
            "downvotes": 2
          }
        },
        // More posts...
      ]
    }
    ```
  - **Error (400)**: Bad Request (invalid pagination parameters)
  - **Error (500)**: Internal Server Error

### Get Post by ID

Retrieves a specific post by its ID.

- **URL**: `/api/v1/posts/{postId}`
- **Method**: `GET`
- **Authentication**: Optional
- **URL Parameters**:
  - `postId`: UUID of the post
- **Response**:
  - **Success (200)**:
    ```json
    {
      "message": "post retrieved successfully",
      "data": {
        "id": "4fa85f64-5717-4562-b3fc-2c963f66afa6",
        "title": "My First Post",
        "content": "This is the content of my first post.",
        "created_at": "2023-04-01T12:00:00Z",
        "updated_at": "2023-04-01T12:00:00Z",
        "categories": ["technology", "golang"],
        "user": {
          "username": "johndoe"
        },
        "vote_count": {
          "upvotes": 10,
          "downvotes": 2
        }
      }
    }
    ```
  - **Error (400)**: Bad Request (invalid post ID format)
  - **Error (404)**: Not Found (post not found)
  - **Error (500)**: Internal Server Error

### Update Post

Updates an existing post.

- **URL**: `/api/v1/posts/{postId}`
- **Method**: `PUT`
- **Authentication**: Required
- **URL Parameters**:
  - `postId`: UUID of the post
- **Request Body**:
  ```json
  {
    "title": "Updated Title",
    "content": "This is the updated content.",
    "categories": ["updated-category"]
  }
  ```
- **Response**:
  - **Success (200)**:
    ```json
    {
      "message": "post updated successfully",
      "data": {
        "id": "4fa85f64-5717-4562-b3fc-2c963f66afa6"
      }
    }
    ```
  - **Error (400)**: Bad Request (validation error or invalid request body)
  - **Error (401)**: Unauthorized (user not logged in)
  - **Error (403)**: Forbidden (user not authorized to update this post)
  - **Error (404)**: Not Found (post not found)
  - **Error (500)**: Internal Server Error

### Vote on Post

Adds or removes a vote (upvote or downvote) on a post.

- **URL**: `/api/v1/posts/{postId}/vote`
- **Method**: `POST`
- **Authentication**: Required
- **URL Parameters**:
  - `postId`: UUID of the post
- **Request Body**:
  ```json
  {
    "vote_type": 1  // 1 for upvote, -1 for downvote
  }
  ```
- **Response**:
  - **Success - Vote Added (200)**:
    ```json
    {
      "message": "vote recorded successfully",
      "data": {
        "id": "4fa85f64-5717-4562-b3fc-2c963f66afa6",
        "vote_count": {
          "upvotes": 11,
          "downvotes": 2
        }
      }
    }
    ```
  - **Success - Vote Removed (200)**:
    ```json
    {
      "message": "successfully removed vote",
      "data": {
        "id": "4fa85f64-5717-4562-b3fc-2c963f66afa6",
        "vote_count": {
          "upvotes": 10,
          "downvotes": 2
        }
      }
    }
    ```
  - **Error (400)**: Bad Request (invalid vote type or request body)
  - **Error (401)**: Unauthorized (user not logged in)
  - **Error (404)**: Not Found (post not found)
  - **Error (500)**: Internal Server Error

## Voting Behavior

- If a user votes with the same vote type they previously used, their vote is removed (toggle behavior)
- If a user votes with a different vote type than their previous vote, their previous vote is replaced with the new one
- A user can have at most one vote per post at any time

## Categories

Posts must belong to at least one category. The categories are predefined in the system and must exist before being associated with a post.