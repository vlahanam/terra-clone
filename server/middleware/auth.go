package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/vlahanam/terra-clone/common"
	"github.com/vlahanam/terra-clone/repositories"
)

const (
	ContextUserIDKey   = "user_id"
	ContextPublicIDKey = "public_id"
	ContextEmailKey    = "email"
	ContextClaimsKey   = "claims"
)

// RequireAuth middleware kiểm tra Header Authorization: Bearer <token> và xác thực token_version
func RequireAuth(jwtSecret string, credRepo repositories.CredentialRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			errResp := common.ErrUnauthorized.
				WithReason("Thiếu Header Authorization").
				WithDetail("header", "Authorization")
			c.AbortWithStatusJSON(http.StatusUnauthorized, errResp)
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			errResp := common.ErrUnauthorized.
				WithReason("Định dạng Authorization Header không hợp lệ, yêu cầu: Bearer <token>")
			c.AbortWithStatusJSON(http.StatusUnauthorized, errResp)
			return
		}

		tokenString := parts[1]
		claims, err := common.ValidateToken(jwtSecret, tokenString, common.TokenTypeAccess)
		if err != nil {
			errResp := common.ErrUnauthorized.
				WithReason("Token không hợp lệ hoặc đã hết hạn").
				WithTrace(err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, errResp)
			return
		}

		// Kiểm tra token_version với cơ sở dữ liệu nếu có CredentialRepository
		if credRepo != nil {
			creds, err := credRepo.FindByUserID(c.Request.Context(), claims.UserID)
			if err != nil || creds == nil || creds.TokenVersion != claims.TokenVersion {
				errResp := common.ErrUnauthorized.
					WithReason("Phiên làm việc đã bị thu hồi hoặc đã đổi mật khẩu, vui lòng đăng nhập lại")
				c.AbortWithStatusJSON(http.StatusUnauthorized, errResp)
				return
			}
		}

		c.Set(ContextUserIDKey, claims.UserID)
		c.Set(ContextPublicIDKey, claims.PublicID)
		c.Set(ContextEmailKey, claims.Email)
		c.Set(ContextClaimsKey, claims)

		c.Next()
	}
}

// GetCurrentUserID trích xuất UserID từ Gin Context
func GetCurrentUserID(c *gin.Context) (uuid.UUID, error) {
	val, exists := c.Get(ContextUserIDKey)
	if !exists {
		return uuid.Nil, errors.New("không tìm thấy user_id trong ngữ cảnh request")
	}

	userID, ok := val.(uuid.UUID)
	if !ok {
		return uuid.Nil, errors.New("user_id sai kiểu dữ liệu")
	}

	return userID, nil
}
