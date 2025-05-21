package main

import (
	"encoding/json"
	"net/http"
	"time"
	"strings"

	"github.com/google/uuid"
	"github.com/vanzei/goserver/internal/auth"
    "github.com/vanzei/goserver/internal/database" 
)

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
	Password string    `json:"password"`
}


type UserResponse struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
	Token	 string    `json:"token,omitempty"`
	IsChirpyRed  bool      `json:"is_chirpy_red"`
}

func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	// Parse the request body
	var req struct {
		Email string `json:"email"`
		Password string `json:"password"`
	}

	type response struct {
		UserResponse
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't decode parameters", err)
		return
	}

	// Validate the email format
	if !isValidEmail(req.Email) {
		respondWithError(w, http.StatusBadRequest, "Invalid email format", nil)
		return
	}

	// Validate the email format
	if !isValidPassword(req.Password) {
		respondWithError(w, http.StatusBadRequest, "Invalid password", nil)
		return
	}
	// Hash the password
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't hash password", err)
		return
	}

	

	// Create a new user in the database
	user, err := cfg.DB.CreateUser(r.Context(), database.CreateUserParams{
        Email:          req.Email,
        HashedPassword: hashedPassword,
    })
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create user", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, response{
		UserResponse: UserResponse{
			ID:        user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Email:     user.Email,
			IsChirpyRed: user.IsChirpyRed,
		},
	})
}



func (cfg *apiConfig) handlerModifyUser(w http.ResponseWriter, r *http.Request) {
    // Get token from Authorization header instead of body
    authHeader := r.Header.Get("Authorization")
    tokenString, err := auth.GetBearerToken(authHeader)
    if err != nil {
        respondWithError(w, http.StatusUnauthorized, "Invalid authorization header", err)
        return
    }
    
    // Validate the token
    userID, err := auth.ValidateJWT(tokenString, cfg.secret)
    if err != nil {
        respondWithError(w, http.StatusUnauthorized, "Invalid token", err)
        return
    }
    
    // Parse the request body (without token field)
    var req struct {
        Email    string `json:"email"`
        Password string `json:"password"`
    }
    
    type response struct {
		UserResponse
	}

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        respondWithError(w, http.StatusBadRequest, "Couldn't decode parameters", err)
        return
    }

    // Validate the email format
    if !isValidEmail(req.Email) {
        respondWithError(w, http.StatusBadRequest, "Invalid email format", nil)
        return
    }

    // Validate the email format
    if !isValidPassword(req.Password) {
        respondWithError(w, http.StatusBadRequest, "Invalid password", nil)
        return
    }

    // Hash the password
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {		
		respondWithError(w, http.StatusInternalServerError, "Couldn't hash password", err)
		return
	}
	// Update the user in the database
	user, err := cfg.DB.UpdateUser(r.Context(), database.UpdateUserParams{
		ID:             userID,
		Email:          req.Email,
		HashedPassword: hashedPassword,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't update user", err)
		return
	}
	respondWithJSON(w, http.StatusOK, response{
		UserResponse: UserResponse{
			ID:        user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Email:     user.Email,
			IsChirpyRed: user.IsChirpyRed,
		},
	})

	
}



func isValidEmail(email string) bool {
	// Simple email validation logic
	if !strings.Contains(email, "@") {
		return false
	}
	return true
}

func isValidPassword(password string) bool {
	// Simple password validation logic
	if len(password) < 1 {
		return false
	}
	return true
}