package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// [VN] RequestIDHeader là tên HTTP Header chứa Request ID
const RequestIDHeader = "X-Request-ID"

// [VN] RequestIDKey là key lưu trữ Request ID trong Gin Context
const RequestIDKey = "request_id"

// [VN] RequestID middleware tự động gán hoặc phát sinh UUIDv7 cho mỗi HTTP request
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		reqID := c.GetHeader(RequestIDHeader)
		if reqID == "" {
			reqID = uuid.Must(uuid.NewV7()).String()
		}

		c.Set(RequestIDKey, reqID)
		c.Header(RequestIDHeader, reqID)
		c.Next()
	}
}
