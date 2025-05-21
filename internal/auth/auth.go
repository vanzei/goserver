package auth

import (
	"net/http"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"time"
	"strings"

)

func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	claims := jwt.RegisteredClaims{
		Issuer:    "chirpy",
		IssuedAt: jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
		Subject:   userID.String(),

	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(tokenSecret))
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(tokenSecret), nil
	})
	if err != nil {
		return uuid.Nil, err
	}
	
	if claims, ok := token.Claims.(*jwt.RegisteredClaims); ok && token.Valid {
		return uuid.Parse(claims.Subject)
	}
	
	return uuid.Nil, fmt.Errorf("invalid token")
}

func GetBearerToken(authHeader string) (string, error) {
    if authHeader == "" {
        return "", fmt.Errorf("missing authorization header")
    }
    
    parts := strings.SplitN(authHeader, " ", 2)
    // Output everything to help debug
    
    if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
        return "", fmt.Errorf("invalid authorization header format")
    }
    
    token := parts[1]
    // Debug check what type of token it might be
    
    return token, nil
}

func MakeRefreshToken() (string, error) {
	// Create a byte slice to hold the random data
	randomBytes := make([]byte, 32)
	
	// Read random bytes - rand.Read returns (n, error)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	
	// Encode as hex string
	encodedToken := hex.EncodeToString(randomBytes)
	return encodedToken, nil
}

func GetAPIKey(headers http.Header) (string, error) {
    authHeader := headers.Get("Authorization")
    if authHeader == "" {
        return "", fmt.Errorf("missing authorization header")
    }
    
    parts := strings.SplitN(authHeader, " ", 2)
    if len(parts) != 2 || strings.ToLower(parts[0]) != "apikey" {
        return "", fmt.Errorf("invalid authorization header format")
    }
    
    return parts[1], nil
}