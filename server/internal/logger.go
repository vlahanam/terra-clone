package internal

import (
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/natefinch/lumberjack/v3"
)

const (
	logFilePath    = "logs/app.log"
	maxSizeBytes   = 10 * 1024 * 1024    // [VN] 10MB
	maxBackups     = 5                   // [VN] Tối đa 5 file backup
	maxAgeDuration = 28 * 24 * time.Hour // [VN] Giữ log trong 28 ngày
	compressOld    = true                // [VN] Nén file cũ dạng .gz
)

// [VN] InitLogger khởi tạo slog ghi log ra console/file và xoay vòng file log tự động.
//
// [VN] Trả về hàm cleanup() chứa fileRoller.Close() nhằm flush toàn bộ log buffer trong RAM xuống đĩa cứng khi shutdown.
// [VN] Cần gọi qua defer để tránh mất log
func InitLogger(c *AppConfig) (func() error, error) {
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
		panic("[VN] Không thể khởi tạo log roller: " + err.Error())
	}

	env := c.AppEnv

	// [VN] Ghi ra cả Console (Stdout) và File
	writer := io.MultiWriter(os.Stdout, fileRoller)
	logLevel := slog.LevelInfo

	if env == "staging" || env == "production" {
		// [VN] Ghi ra File
		writer = fileRoller
		logLevel = slog.LevelWarn
	}

	opts := &slog.HandlerOptions{
		Level:     logLevel,
		AddSource: true,
	}

	handler := slog.NewJSONHandler(writer, opts)

	log := slog.New(handler)
	slog.SetDefault(log)

	cleanup := func() error {
		return fileRoller.Close()
	}

	return cleanup, nil
}
