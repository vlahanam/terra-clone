package middleware

import (
	"fmt"
	"io"
	"time"

	"github.com/gin-gonic/gin"
)

// [VN] Logger cấu hình middleware ghi access log cho Gin kèm thông tin Client IP và Thiết bị (User-Agent) vào output được chỉ định.
func Logger(output io.Writer) gin.HandlerFunc {
	if output == nil {
		output = gin.DefaultWriter
	}

	return gin.LoggerWithConfig(gin.LoggerConfig{
		Output: output,
		Formatter: func(param gin.LogFormatterParams) string {
			var statusColor, methodColor, resetColor string
			if param.IsOutputColor() {
				statusColor = param.StatusCodeColor()
				methodColor = param.MethodColor()
				resetColor = param.ResetColor()
			}

			if param.Latency > time.Minute {
				param.Latency = param.Latency.Truncate(time.Second)
			}

			clientIP := param.ClientIP
			if clientIP == "" {
				clientIP = "-"
			}

			device := param.Request.UserAgent()
			if device == "" {
				device = "Unknown"
			}

			return fmt.Sprintf("[ACCESS] %v |%s %3d %s| %13v | IP: %15s |%s %-7s %s %#v | Device: %s\n%s",
				param.TimeStamp.Format("2006/01/02 - 15:04:05"),
				statusColor, param.StatusCode, resetColor,
				param.Latency,
				clientIP,
				methodColor, param.Method, resetColor,
				param.Path,
				device,
				param.ErrorMessage,
			)
		},
	})
}

