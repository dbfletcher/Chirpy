package auth_test

import (
	"testing"
	"time"

	"github.com/dbfletcher/chirpy/internal/auth"
	"github.com/google/uuid"
)

func TestJWTFunctions(t *testing.T) {
	const testSecret = "my-super-secret-key-for-testing"

	t.Run("ValidToken", func(t *testing.T) {
		userID := uuid.New()
		tokenString, err := auth.MakeJWT(userID, testSecret, time.Hour)
		if err != nil {
			t.Fatalf("Failed to create valid JWT: %v", err)
		}

		validatedUserID, err := auth.ValidateJWT(tokenString, testSecret)
		if err != nil {
			t.Fatalf("Failed to validate a valid JWT: %v", err)
		}

		if validatedUserID != userID {
			t.Errorf("Expected user ID %v, but got %v", userID, validatedUserID)
		}
	})

	t.Run("ExpiredToken", func(t *testing.T) {
		userID := uuid.New()
		// Create a token that expired an hour ago
		tokenString, err := auth.MakeJWT(userID, testSecret, -1*time.Hour)
		if err != nil {
			t.Fatalf("Failed to create expired JWT: %v", err)
		}

		_, err = auth.ValidateJWT(tokenString, testSecret)
		if err == nil {
			t.Error("Expected an error for an expired token, but got none")
		}
	})

	t.Run("WrongSecret", func(t *testing.T) {
		userID := uuid.New()
		tokenString, err := auth.MakeJWT(userID, testSecret, time.Hour)
		if err != nil {
			t.Fatalf("Failed to create JWT: %v", err)
		}

		_, err = auth.ValidateJWT(tokenString, "a-different-secret")
		if err == nil {
			t.Error("Expected an error for a token with the wrong secret, but got none")
		}
	})
}
