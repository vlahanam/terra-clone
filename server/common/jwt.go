package common

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"

	AccessTokenDuration  = 15 * time.Minute
	RefreshTokenDuration = 7 * 24 * time.Hour
)

type UserClaims struct {
	UserID       uuid.UUID `json:"user_id"`
	PublicID     string    `json:"public_id"`
	Email        string    `json:"email"`
	TokenVersion int       `json:"token_version"`
	TokenType    string    `json:"token_type"`
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"` // seconds
}

func GenerateTokenPair(secret string, userID uuid.UUID, publicID, email string, tokenVersion int) (*TokenPair, error) {
	now := time.Now()

	// Access Token
	accessClaims := UserClaims{
		UserID:       userID,
		PublicID:     publicID,
		Email:        email,
		TokenVersion: tokenVersion,
		TokenType:    TokenTypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(AccessTokenDuration)),
		},
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	signedAccess, err := accessToken.SignedString([]byte(secret))
	if err != nil {
		return nil, fmt.Errorf("lỗi ký access token: %w", err)
	}

	// Refresh Token
	refreshClaims := UserClaims{
		UserID:       userID,
		PublicID:     publicID,
		Email:        email,
		TokenVersion: tokenVersion,
		TokenType:    TokenTypeRefresh,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(RefreshTokenDuration)),
		},
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	signedRefresh, err := refreshToken.SignedString([]byte(secret))
	if err != nil {
		return nil, fmt.Errorf("lỗi ký refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  signedAccess,
		RefreshToken: signedRefresh,
		TokenType:    "Bearer",
		ExpiresIn:    int64(AccessTokenDuration.Seconds()),
	}, nil
}

func ValidateToken(secret, tokenString, expectedType string) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("phương thức ký không được hỗ trợ: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*UserClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("token không hợp lệ")
	}

	if expectedType != "" && claims.TokenType != expectedType {
		return nil, fmt.Errorf("loại token không khớp: mong đợi %s nhưng nhận %s", expectedType, claims.TokenType)
	}

	return claims, nil
}
