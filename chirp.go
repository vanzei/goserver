package main

import (
	"net/http"
	"encoding/json"
	"strings"
    "time"
    "errors"
	"sort"

	"github.com/google/uuid"
    "github.com/vanzei/goserver/internal/auth"
    "github.com/vanzei/goserver/internal/database"
)

type ChirpResponse struct {
    ID        uuid.UUID  `json:"id"`
    CreatedAt time.Time  `json:"created_at"`
    UpdatedAt time.Time  `json:"updated_at"`
    Body      string     `json:"body"`
    UserID    uuid.UUID  `json:"user_id"`
}

// Helper function to extract and validate JWT token
func (cfg *apiConfig) validateJWTFromRequest(r *http.Request) (uuid.UUID, error) {
    authHeader := r.Header.Get("Authorization")
    if authHeader == "" {
        return uuid.UUID{}, ErrMissingAuthHeader
    }
    
    parts := strings.SplitN(authHeader, " ", 2)
    if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
        return uuid.UUID{}, ErrInvalidAuthHeaderFormat
    }
    
    userID, err := auth.ValidateJWT(parts[1], cfg.secret)
    if err != nil {
        return uuid.UUID{}, err
    }
    
    return userID, nil
}

// Helper function to extract chirp ID from request path
func getChirpIDFromPath(r *http.Request) (uuid.UUID, error) {
    chirpIDStr := r.PathValue("chirpID")
    if chirpIDStr == "" {
        return uuid.UUID{}, ErrMissingChirpID
    }
    
    chirpID, err := uuid.Parse(chirpIDStr)
    if err != nil {
        return uuid.UUID{}, err
    }
    
    return chirpID, nil
}

// Define custom errors for better handling
var (
    ErrMissingAuthHeader      = errors.New("missing authorization header")
    ErrInvalidAuthHeaderFormat = errors.New("invalid authorization header format")
    ErrMissingChirpID         = errors.New("chirp ID is required")
)

func (cfg *apiConfig) handlerCreateChirp(w http.ResponseWriter, r *http.Request) {
    type parameters struct {
        Body string `json:"body"`
        UserID uuid.UUID `json:"user_id"`
    }
    
    decoder := json.NewDecoder(r.Body)
    params := parameters{}
    err := decoder.Decode(&params)
    if err != nil {
        respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
        return
    }

    const maxChirpLength = 140
    if len(params.Body) > maxChirpLength {
        respondWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
        return
    }

    // Use helper function to validate JWT
    userID, err := cfg.validateJWTFromRequest(r)
    if err != nil {
        switch err {
        case ErrMissingAuthHeader:
            respondWithError(w, http.StatusUnauthorized, "Missing authorization header", nil)
        case ErrInvalidAuthHeaderFormat:
            respondWithError(w, http.StatusUnauthorized, "Invalid authorization header format", nil)
        default:
            respondWithError(w, http.StatusUnauthorized, "Invalid token", err)
        }
        return
    }

    // Process text to find profane words (case insensitive)
    cleanedBody := params.Body
    lowerText := strings.ToLower(params.Body)
    
    profaneWords := []string{"kerfuffle", "sharbert", "fornax"}
    for _, profaneWord := range profaneWords {
        // Find all instances of the profane word (case insensitive)
        index := strings.Index(lowerText, profaneWord)
        for index != -1 {
            // Replace in the original text while preserving case
            cleanedBody = cleanedBody[:index] + "****" + cleanedBody[index+len(profaneWord):]
            // Also update the lowercase text for further searches
            lowerText = lowerText[:index] + "****" + lowerText[index+len(profaneWord):]
            // Find the next instance
            index = strings.Index(lowerText, profaneWord)
        }
    }

    // Create the chirp in the database
    chirp, err := cfg.DB.CreateChirp(r.Context(), database.CreateChirpParams{
        Body: cleanedBody,
        UserID: uuid.NullUUID{
        UUID:  userID,  // Use the ID from the token
        Valid: true,
    },
    })

    if err != nil {
        respondWithError(w, http.StatusInternalServerError, "Couldn't create chirp", err)
        return
    }
    
    // Respond with the created chirp
    respondWithJSON(w, http.StatusCreated, ChirpResponse{
        ID:        chirp.ID,
        CreatedAt: chirp.CreatedAt,
        UpdatedAt: chirp.UpdatedAt,
        Body:      chirp.Body,
        UserID:    chirp.UserID.UUID,
    })
}

func (cfg *apiConfig) handlerGetChirps(w http.ResponseWriter, r *http.Request) {
    // Check if author_id query parameter exists
    authorIDStr := r.URL.Query().Get("author_id")
    
    // Get sort direction from query parameter (default is "asc")
    sortDir := r.URL.Query().Get("sort")
    if sortDir != "desc" && sortDir != "asc" {
        sortDir = "asc" // Default to ascending if invalid or missing
    }
    
    var chirps []database.Chirp
    var err error
    
    if authorIDStr != "" {
        // Parse author ID
        authorID, err := uuid.Parse(authorIDStr)
        if err != nil {
            respondWithError(w, http.StatusBadRequest, "Invalid author ID format", err)
            return
        }
        
        // Get chirps filtered by author
        chirps, err = cfg.DB.GetChirpsByAuthor(r.Context(), uuid.NullUUID{
            UUID:  authorID,
            Valid: true,
        })
    } else {
        // Get all chirps
        chirps, err = cfg.DB.GetChirps(r.Context())
    }
    
    if err != nil {
        respondWithError(w, http.StatusInternalServerError, "Couldn't get chirps", err)
        return
    }
    
    // Sort chirps in memory based on the sort parameter
    sort.Slice(chirps, func(i, j int) bool {
        if sortDir == "asc" {
            return chirps[i].CreatedAt.Before(chirps[j].CreatedAt)
        }
        return chirps[i].CreatedAt.After(chirps[j].CreatedAt)
    })
    
    // Convert database chirps to response chirps
    chirpResponses := []ChirpResponse{}
    for _, chirp := range chirps {
        chirpResponses = append(chirpResponses, ChirpResponse{
            ID:        chirp.ID,
            CreatedAt: chirp.CreatedAt,
            UpdatedAt: chirp.UpdatedAt,
            Body:      chirp.Body,
            UserID:    chirp.UserID.UUID,
        })
    }
    
    respondWithJSON(w, http.StatusOK, chirpResponses)
}

func (cfg *apiConfig) handlerGetChirpbyId(w http.ResponseWriter, r *http.Request) {
    // Use helper function to get chirpID
    chirpID, err := getChirpIDFromPath(r)
    if err != nil {
        switch err {
        case ErrMissingChirpID:
            respondWithError(w, http.StatusNotFound, "Chirp ID is required", nil)
        default:
            respondWithError(w, http.StatusBadRequest, "Invalid chirp ID format", err)
        }
        return
    }

    chirp, err := cfg.DB.GetChirpbyId(r.Context(), chirpID)
    if err != nil {
        respondWithError(w, http.StatusNotFound, "Couldn't get chirp", err)
        return
    }

    respondWithJSON(w, http.StatusOK, ChirpResponse{
        ID:        chirp.ID,
        CreatedAt: chirp.CreatedAt,
        UpdatedAt: chirp.UpdatedAt,
        Body:      chirp.Body,
        UserID:    chirp.UserID.UUID,
    })
}

func (cfg *apiConfig) handlerDeleteChirpbyId(w http.ResponseWriter, r *http.Request) {
    // Use helper function to validate JWT
    userID, err := cfg.validateJWTFromRequest(r)
    if err != nil {
        switch err {
        case ErrMissingAuthHeader:
            respondWithError(w, http.StatusUnauthorized, "Missing authorization header", nil)
        case ErrInvalidAuthHeaderFormat:
            respondWithError(w, http.StatusUnauthorized, "Invalid authorization header format", nil)
        default:
            respondWithError(w, http.StatusUnauthorized, "Invalid token", err)
        }
        return
    }

    // Use helper function to get chirpID
    chirpID, err := getChirpIDFromPath(r)
    if err != nil {
        switch err {
        case ErrMissingChirpID:
            respondWithError(w, http.StatusNotFound, "Chirp ID is required", nil)
        default:
            respondWithError(w, http.StatusBadRequest, "Invalid chirp ID format", err)
        }
        return
    }

    // check if the chirp belongs to the user
    chirp, err := cfg.DB.GetChirpbyId(r.Context(), chirpID)
    if err != nil {
        respondWithError(w, http.StatusNotFound, "Couldn't get chirp", err)
        return
    }

    if chirp.UserID.UUID != userID {
        respondWithError(w, http.StatusForbidden, "You are not authorized to delete this chirp", nil)
        return
    }

    err = cfg.DB.DeleteChirpbyId(r.Context(), chirpID)
    if err != nil {
        respondWithError(w, http.StatusInternalServerError, "Couldn't delete chirp", err)
        return
    }

    respondWithJSON(w, http.StatusNoContent, nil)
}