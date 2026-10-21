package common

import (
	"testing"

	"github.com/google/uuid"
)

func TestGenerateAndValidateToken(t *testing.T) {
	secret := "test_secret_key_1234567890_min_32_bytes!"
	userID := uuid.New()
	publicID := "usr_test123"
	email := "test@example.com"
	tokenVersion := 1

	pair, err := GenerateTokenPair(secret, userID, publicID, email, tokenVersion)
	if err != nil {
		t.Fatalf("GenerateTokenPair failed: %v", err)
	}

	if pair.AccessToken == "" || pair.RefreshToken == "" {
		t.Fatalf("Expected non-empty tokens, got access=%s, refresh=%s", pair.AccessToken, pair.RefreshToken)
	}

	// Validate Access Token
	accessClaims, err := ValidateToken(secret, pair.AccessToken, TokenTypeAccess)
	if err != nil {
		t.Fatalf("ValidateToken for access failed: %v", err)
	}
	if accessClaims.UserID != userID || accessClaims.Email != email || accessClaims.TokenVersion != tokenVersion {
		t.Errorf("Access claims mismatch: %+v", accessClaims)
	}

	// Validate Refresh Token
	refreshClaims, err := ValidateToken(secret, pair.RefreshToken, TokenTypeRefresh)
	if err != nil {
		t.Fatalf("ValidateToken for refresh failed: %v", err)
	}
	if refreshClaims.UserID != userID || refreshClaims.TokenVersion != tokenVersion {
		t.Errorf("Refresh claims mismatch: %+v", refreshClaims)
	}

	// Validate wrong token type
	if _, err := ValidateToken(secret, pair.AccessToken, TokenTypeRefresh); err == nil {
		t.Errorf("Expected error when validating access token as refresh token, got nil")
	}

	// Validate wrong secret
	if _, err := ValidateToken("wrong_secret_key", pair.AccessToken, TokenTypeAccess); err == nil {
		t.Errorf("Expected error when validating with wrong secret, got nil")
	}
}

func TestGeneratePublicID(t *testing.T) {
	id := GeneratePublicID("usr")
	if len(id) < 5 || id[:4] != "usr_" {
		t.Errorf("Unexpected public ID format: %s", id)
	}
}
