package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/vanzei/goserver/internal/auth"
	"github.com/vanzei/goserver/internal/database"
)

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	// Parse the request body
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

	// find email in db
	user, err := cfg.DB.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusUnauthorized, "Invalid email or password", nil)
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Couldn't find user", err)
		return
	}
	// Check if the password is correct
	if !auth.CheckPasswordHash(req.Password, user.HashedPassword) {
		respondWithError(w, http.StatusUnauthorized, "Invalid email or password", nil)
		return
	}

	// Generate access token with fixed 1-hour expiration
	accessToken, err := auth.MakeJWT(user.ID, cfg.secret, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't generate access token", err)
		return
	}

	// Generate refresh token
	refreshTokenString, err := auth.MakeRefreshToken()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't generate refresh token", err)
		return
	}

	// Store refresh token in database - use _ to ignore the variable if you don't need it
	_, err = cfg.DB.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		UserID: user.ID,
		Token:  refreshTokenString,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't store refresh token", err)
		return
	}

	// Return both tokens in response
	respondWithJSON(w, http.StatusOK, struct {
		ID            uuid.UUID `json:"id"`
		Email         string    `json:"email"`
		Token         string    `json:"token"`
		RefreshToken  string    `json:"refresh_token"`
		IsChirpyRed  bool      `json:"is_chirpy_red"`
	}{
		ID:           user.ID,
		Email:        user.Email,
		Token:        accessToken,
		RefreshToken: refreshTokenString,
		IsChirpyRed: user.IsChirpyRed,
	})
}

