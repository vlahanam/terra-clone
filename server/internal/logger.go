package internal

import (
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/natefinch/lumberjack/v3"
)

const (
	logFilePath       = "logs/app.log"
	accessLogFilePath = "logs/access.log"
	maxSizeBytes      = 10 * 1024 * 1024    // [VN] 10MB
	maxBackups        = 5                   // [VN] Tối đa 5 file backup
	maxAgeDuration    = 28 * 24 * time.Hour // [VN] Giữ log trong 28 ngày
	compressOld       = true                // [VN] Nén file cũ dạng .gz
)

// [VN] InitLogger khởi tạo slog ghi log ứng dụng ra app.log và access log ra access.log với cơ chế xoay vòng tự động.
//
// [VN] Trả về accessWriter cho router và hàm cleanup() đóng toàn bộ log rollers khi shutdown.
// [VN] Cần gọi qua defer để tránh mất log
func InitLogger(c *AppConfig) (io.Writer, func() error, error) {
	fileRoller, err := lumberjack.NewRoller(
		logFilePath,
		maxSizeBytes,
		&lumberjack.Options{
			MaxBackups: maxBackups,
			MaxAge:     maxAgeDuration,
			Compress:   compressOld,
		},
	)
	if err != nil {
		panic("failed to initialize application log roller: " + err.Error())
	}

	accessRoller, err := lumberjack.NewRoller(
		accessLogFilePath,
		maxSizeBytes,
		&lumberjack.Options{
			MaxBackups: maxBackups,
			MaxAge:     maxAgeDuration,
			Compress:   compressOld,
		},
	)
	if err != nil {
		panic("failed to initialize access log roller: " + err.Error())
	}

	env := c.AppEnv

	// [VN] Ghi ra cả Console (Stdout) và File
	writer := io.MultiWriter(os.Stdout, fileRoller)
	accessWriter := io.MultiWriter(os.Stdout, accessRoller)
	logLevel := slog.LevelInfo

	if env == "staging" || env == "production" {
		// [VN] Ghi ra File
		writer = fileRoller
		accessWriter = accessRoller
		logLevel = slog.LevelWarn
	}

	// [VN] Chuyển hướng log lỗi của GIN Framework sang app writer
	gin.DefaultWriter = writer
	gin.DefaultErrorWriter = writer

	opts := &slog.HandlerOptions{
		Level:     logLevel,
		AddSource: true,
	}

	handler := slog.NewJSONHandler(writer, opts)

	log := slog.New(handler)
	slog.SetDefault(log)

	cleanup := func() error {
		err1 := fileRoller.Close()
		err2 := accessRoller.Close()
		if err1 != nil {
			return err1
		}
		return err2
	}

	return accessWriter, cleanup, nil
}
