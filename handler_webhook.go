package main

import (
    "encoding/json"
    "net/http"
    "database/sql"

    "github.com/google/uuid"
	"github.com/vanzei/goserver/internal/auth"
	"github.com/vanzei/goserver/internal/database"
)

func (cfg *apiConfig) handlerWebhook(w http.ResponseWriter, r *http.Request) {

	// Validate API key
    apiKey, err := auth.GetAPIKey(r.Header)
    if err != nil {
        respondWithError(w, http.StatusUnauthorized, "Unauthorized", err)
        return
    }
    
    // Check if API key matches
    if apiKey != cfg.polkaWebhookSecret {
        respondWithError(w, http.StatusUnauthorized, "Invalid API key", nil)
        return
    }

    // Parse the request body
    var req struct {
        Event string `json:"event"`
        Data  struct {
            UserID string `json:"user_id"`
        } `json:"data"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        respondWithError(w, http.StatusBadRequest, "Couldn't decode request body", err)
        return
    }

    // If event is not user.upgraded, return 204 immediately
    if req.Event != "user.upgraded" {
        w.WriteHeader(http.StatusNoContent)
        return
    }

    // Parse the user ID
    userID, err := uuid.Parse(req.Data.UserID)
    if err != nil {
        respondWithError(w, http.StatusBadRequest, "Invalid user ID format", err)
        return
    }

    // Update the user's Chirpy Red status
    _, err = cfg.DB.UpdateUserChirpyRed(r.Context(), database.UpdateUserChirpyRedParams{
        ID:          userID,
        IsChirpyRed: true,
    })
    
    if err != nil {
        if err == sql.ErrNoRows {
            respondWithError(w, http.StatusNotFound, "User not found", nil)
        } else {
            respondWithError(w, http.StatusInternalServerError, "Couldn't update user", err)
        }
        return
    }

    // Return 204 No Content on success
    w.WriteHeader(http.StatusNoContent)
}