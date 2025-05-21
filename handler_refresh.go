package main

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/vanzei/goserver/internal/auth"
)

func (cfg *apiConfig) handlerRefresh(w http.ResponseWriter, r *http.Request) {
	type response struct {
		Token string `json:"token"`
	}

	// Get the refresh token from Authorization header
	authHeader := r.Header.Get("Authorization")
	refreshToken, err := auth.GetBearerToken(authHeader)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid authorization header", err)
		return
	}
	// Look up the token in the database
	user, err := cfg.DB.GetUserFromRefreshToken(r.Context(), refreshToken)
	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusUnauthorized, "Invalid or expired refresh token", nil)
		} else {
			respondWithError(w, http.StatusInternalServerError, "Couldn't validate refresh token", err)
		}
		return
	}
	accessToken, err := auth.MakeJWT(
		user.ID,
		cfg.secret,
		time.Hour,
	)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate token", err)
		return
	}

	respondWithJSON(w, http.StatusOK, response{
		Token: accessToken,
	})
}

func (cfg *apiConfig) handlerRevoke(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header.Get("Authorization"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't find token", err)
		return
	}

	_, err = cfg.DB.RevokeRefreshToken(r.Context(), refreshToken)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't revoke session", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}