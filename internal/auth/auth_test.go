package auth

import (
    "strings"
    "testing"
    "time"

    "github.com/google/uuid"
)

func TestHashPassword(t *testing.T) {
    // Test 1: Valid password
    t.Run("Valid password", func(t *testing.T) {
        password := "SecureP@ssw0rd123"
        hash, err := HashPassword(password)
        if err != nil {
            t.Fatalf("HashPassword returned error: %v", err)
        }
        if hash == "" {
            t.Fatal("HashPassword returned empty hash")
        }
        if hash == password {
            t.Fatal("Hash should not be equal to original password")
        }
        // bcrypt hashes start with $2a$, $2b$ or $2y$
        if !strings.HasPrefix(hash, "$2") {
            t.Fatalf("Hash does not have bcrypt prefix: %s", hash)
        }
    })

    // Test 2: Empty password (edge case but should work)
    t.Run("Empty password", func(t *testing.T) {
        hash, err := HashPassword("")
        if err != nil {
            t.Fatalf("HashPassword with empty string returned error: %v", err)
        }
        if hash == "" {
            t.Fatal("HashPassword returned empty hash")
        }
    })

    // Test 3: Password at maximum length
    t.Run("Maximum length password", func(t *testing.T) {
        // Create a 72 character password (bcrypt's limit)
        maxLengthPassword := strings.Repeat("a", 72)
        hash, err := HashPassword(maxLengthPassword)
        if err != nil {
            t.Fatalf("HashPassword with max length password returned error: %v", err)
        }
        if hash == "" {
            t.Fatal("HashPassword returned empty hash")
        }
    })
}

func TestCheckPasswordHash(t *testing.T) {
    // Test 1: Correct password
    t.Run("Correct password", func(t *testing.T) {
        password := "SecureP@ssw0rd123"
        hash, err := HashPassword(password)
        if err != nil {
            t.Fatalf("HashPassword returned error: %v", err)
        }
        
        if !CheckPasswordHash(password, hash) {
            t.Fatal("CheckPasswordHash should return true for correct password")
        }
    })

    // Test 2: Incorrect password
    t.Run("Incorrect password", func(t *testing.T) {
        password := "SecureP@ssw0rd123"
        wrongPassword := "WrongPassword123"
        hash, err := HashPassword(password)
        if err != nil {
            t.Fatalf("HashPassword returned error: %v", err)
        }
        
        if CheckPasswordHash(wrongPassword, hash) {
            t.Fatal("CheckPasswordHash should return false for incorrect password")
        }
    })

    // Test 3: Invalid hash format
    t.Run("Invalid hash format", func(t *testing.T) {
        password := "SecureP@ssw0rd123"
        invalidHash := "not-a-valid-bcrypt-hash"
        
        if CheckPasswordHash(password, invalidHash) {
            t.Fatal("CheckPasswordHash should return false for invalid hash format")
        }
    })
}

func TestMakeJWT(t *testing.T) {
    // Test 1: Valid JWT creation
    t.Run("Valid JWT creation", func(t *testing.T) {
        userID := uuid.New()
        tokenSecret := "test-secret"
        expiresIn := time.Hour * 24
        
        token, err := MakeJWT(userID, tokenSecret, expiresIn)
        if err != nil {
            t.Fatalf("MakeJWT returned error: %v", err)
        }
        if token == "" {
            t.Fatal("MakeJWT returned empty token")
        }
        
        // JWT should have 3 parts separated by periods
        parts := strings.Split(token, ".")
        if len(parts) != 3 {
            t.Fatalf("Invalid JWT format, expected 3 parts but got %d", len(parts))
        }
    })
    
    // Test 2: Empty secret
    t.Run("Empty secret", func(t *testing.T) {
        userID := uuid.New()
        emptySecret := ""
        expiresIn := time.Hour
        
        token, err := MakeJWT(userID, emptySecret, expiresIn)
        if err != nil {
            t.Fatalf("MakeJWT with empty secret returned error: %v", err)
        }
        if token == "" {
            t.Fatal("MakeJWT returned empty token")
        }
    })
    
    // Test 3: Zero expiration
    t.Run("Zero expiration", func(t *testing.T) {
        userID := uuid.New()
        tokenSecret := "test-secret"
        zeroExpiration := time.Duration(0)
        
        token, err := MakeJWT(userID, tokenSecret, zeroExpiration)
        if err != nil {
            t.Fatalf("MakeJWT with zero expiration returned error: %v", err)
        }
        if token == "" {
            t.Fatal("MakeJWT returned empty token")
        }
    })
}

func TestValidateJWT(t *testing.T) {
    // Test 1: Valid token validation
    t.Run("Valid token", func(t *testing.T) {
        userID := uuid.New()
        tokenSecret := "test-secret"
        expiresIn := time.Hour
        
        token, err := MakeJWT(userID, tokenSecret, expiresIn)
        if err != nil {
            t.Fatalf("MakeJWT returned error: %v", err)
        }
        
        extractedID, err := ValidateJWT(token, tokenSecret)
        if err != nil {
            t.Fatalf("ValidateJWT returned error: %v", err)
        }
        
        if extractedID != userID {
            t.Fatalf("Extracted ID (%s) doesn't match original ID (%s)", extractedID, userID)
        }
    })
    
    // Test 2: Invalid token format
    t.Run("Invalid token format", func(t *testing.T) {
        invalidToken := "invalid.token.format"
        tokenSecret := "test-secret"
        
        _, err := ValidateJWT(invalidToken, tokenSecret)
        if err == nil {
            t.Fatal("ValidateJWT should return error for invalid token")
        }
    })
    
    // Test 3: Wrong secret
    t.Run("Wrong secret", func(t *testing.T) {
        userID := uuid.New()
        tokenSecret := "correct-secret"
        wrongSecret := "wrong-secret"
        expiresIn := time.Hour
        
        token, err := MakeJWT(userID, tokenSecret, expiresIn)
        if err != nil {
            t.Fatalf("MakeJWT returned error: %v", err)
        }
        
        _, err = ValidateJWT(token, wrongSecret)
        if err == nil {
            t.Fatal("ValidateJWT should return error for token with wrong secret")
        }
    })
    
    // Test 4: Expired token
    t.Run("Expired token", func(t *testing.T) {
        userID := uuid.New()
        tokenSecret := "test-secret"
        expiresIn := time.Millisecond // Very short duration
        
        token, err := MakeJWT(userID, tokenSecret, expiresIn)
        if err != nil {
            t.Fatalf("MakeJWT returned error: %v", err)
        }
        
        // Wait for token to expire
        time.Sleep(time.Millisecond * 5)
        
        _, err = ValidateJWT(token, tokenSecret)
        if err == nil {
            t.Fatal("ValidateJWT should return error for expired token")
        }
    })
}