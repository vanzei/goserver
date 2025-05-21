# Chirpy API Server

A RESTful API server for a Twitter-like social media platform built with Go, featuring JWT authentication, refresh tokens, and premium user features.

## Features

- User management (registration, login, profile updates)
- Post short messages called "chirps"
- Content moderation (profanity filter)
- JWT-based authentication with refresh tokens
- Premium user subscriptions (Chirpy Red)
- API metrics and monitoring
- Webhook integration for third-party services

## Setup Instructions

### Prerequisites

- Go 1.22+
- PostgreSQL database
- Git

### Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/vanzei/goserver.git
   cd goserver
2. Install dependencies:
    ```bash
    go mod download
3. Set up environment variables (create a .env file):
    ```bash
    DB_URL=postgres://username:password@localhost:5432/chirpy?sslmode=disable
    PLATFORM=production
    secret=your-jwt-secret-key
    POLKA_KEY=your-polka-webhook-key
4. Run database migrations:
    ```bash
    goose -dir sql/schema up
5. Start the server:
    ```
    go run .
    ```

## Authentication

### User Authentication

The application uses two types of tokens:
- Access Token: JWT token valid for 1 hour, used to authenticate API requests
- Refresh Token: Long-lived token (60 days) used to obtain new access tokens

Authentication headers should be in the format:
```

Authorization: Bearer YOUR_TOKEN_HERE
```


```
Authorization: Bearer YOUR_ACCESS_TOKEN
```

## API Endpoints

### Authentication

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|--------------|
| POST | `/api/users` | Register a new user | No |
| POST | `/api/login` | Login and get tokens | No |
| POST | `/api/refresh` | Get a new access token | Yes (Refresh token) |
| POST | `/api/revoke` | Revoke a refresh token | Yes (Refresh token) |

### Users

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|--------------|
| PUT | `/api/users` | Update user profile | Yes (Access token) |

### Chirps

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|--------------|
| POST | `/api/chirps` | Create a new chirp | Yes (Access token) |
| GET | `/api/chirps` | Get all chirps with optional filtering | No |
| GET | `/api/chirps/{chirpID}` | Get a specific chirp | No |
| DELETE | `/api/chirps/{chirpID}` | Delete a chirp | Yes (Access token, owner only) |

### Webhooks

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|--------------|
| POST | `/api/polka/webhooks` | Handle premium subscription events | Yes (API Key) |

### Admin

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|--------------|
| GET | `/admin/metrics` | Get server metrics | No |
| POST | `/admin/reset` | Reset server metrics | No |



Common HTTP status codes:

* 200: Success
* 201: Resource created
* 204: Success (no content)
* 400: Bad request
* 401: Unauthorized
* 403: Forbidden
* 404: Not found
* 500: Internal server error

### Security Features

* Password hashing with bcrypt
* JWT token validation
* API key validation for webhooks
* User-owned resource authorization

### Premium Features (Chirpy Red)
Users can upgrade to Chirpy Red through the Polka payment service. Premium status is indicated by the ***is_chirpy_red*** field in user responses.