package common

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// [VN] UserClaims lưu trữ thông tin định danh và phân quyền bên trong JWT payload
type UserClaims struct {
	UserID       uuid.UUID `json:"user_id"`
	PublicID     string    `json:"public_id"`
	Email        string    `json:"email"`
	TokenVersion int       `json:"token_version"`
	TokenType    string    `json:"token_type"`
	jwt.RegisteredClaims
}

// [VN] TokenPair đại diện cho cặp token trả về cho client sau khi xác thực thành công
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"` // [VN] Thời hạn access token tính bằng giây
}

// [VN] GenerateTokenPair sinh cặp Access Token và Refresh Token cho người dùng
func GenerateTokenPair(secret string, userID uuid.UUID, publicID, email string, tokenVersion int) (*TokenPair, error) {
	now := time.Now()

	// [VN] Tạo payload cho Access Token
	accessClaims := UserClaims{
		UserID:       userID,
		PublicID:     publicID,
		Email:        email,
		TokenVersion: tokenVersion,
		TokenType:    JWT_TOKEN_TYPE_ACCESS,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(JWT_ACCESS_TOKEN_DURATION)),
		},
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	signedAccess, err := accessToken.SignedString([]byte(secret))
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	// [VN] Tạo payload cho Refresh Token
	refreshClaims := UserClaims{
		UserID:       userID,
		PublicID:     publicID,
		Email:        email,
		TokenVersion: tokenVersion,
		TokenType:    JWT_TOKEN_TYPE_REFRESH,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(JWT_REFRESH_TOKEN_DURATION)),
		},
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	signedRefresh, err := refreshToken.SignedString([]byte(secret))
	if err != nil {
		return nil, fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  signedAccess,
		RefreshToken: signedRefresh,
		TokenType:    "Bearer",
		ExpiresIn:    int64(JWT_ACCESS_TOKEN_DURATION.Seconds()),
	}, nil
}

// [VN] ValidateToken kiểm tra tính hợp lệ, chữ ký và loại token của chuỗi JWT
func ValidateToken(secret, tokenString, expectedType string) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*UserClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	if expectedType != "" && claims.TokenType != expectedType {
		return nil, fmt.Errorf("token type mismatch: expected %s but got %s", expectedType, claims.TokenType)
	}

	return claims, nil
}
