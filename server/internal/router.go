package internal

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vlahanam/terra-clone/controllers"
	"github.com/vlahanam/terra-clone/middleware"
	"github.com/vlahanam/terra-clone/repositories"
	"github.com/vlahanam/terra-clone/services"
	"gorm.io/gorm"
)

// [VN] SetupRouter khởi tạo Gin Engine, gắn middleware và khai báo các route API v1
func SetupRouter(c *AppConfig, db *gorm.DB, accessWriter io.Writer) *gin.Engine {
	if c.AppEnv == "staging" || c.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(
		middleware.RequestID(),
		middleware.Logger(accessWriter),
		gin.Recovery(),
	)

	// [VN] Các endpoint kiểm tra sức khỏe hệ thống (Health Check)
	pingHandler := func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	}
	r.GET("/ping", pingHandler)
	r.GET("/api/ping", pingHandler)

	// [VN] Khởi tạo tầng Repositories qua Dependency Injection
	userRepo := repositories.NewUserRepository(db)
	credRepo := repositories.NewCredentialRepository(db)
	profileRepo := repositories.NewProfileRepository(db)

	// [VN] Khởi tạo tầng Services qua Dependency Injection
	authService := services.NewAuthService(userRepo, credRepo, profileRepo, c.JWTSecret)
	userService := services.NewUserService(userRepo, credRepo, profileRepo)

	// [VN] Khởi tạo tầng Controllers qua Dependency Injection
	authController := controllers.NewAuthController(authService)
	userController := controllers.NewUserController(userService)

	// [VN] Khởi tạo Middleware xác thực JWT
	authMiddleware := middleware.RequireAuth(c.JWTSecret, credRepo)

	// [VN] Định tuyến API v1: /api/v1
	api := r.Group("/api")
	v1 := api.Group("/v1")
	{
		// [VN] Nhóm route xác thực: /api/v1/auth
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authController.Register)
			auth.POST("/login", authController.Login)
			auth.POST("/refresh", authController.RefreshToken)
			auth.GET("/me", authMiddleware, authController.GetMe)
			auth.POST("/logout-all", authMiddleware, authController.LogoutAll)
		}

		// [VN] Nhóm route quản lý người dùng: /api/v1/users
		users := v1.Group("/users")
		{
			users.GET("", userController.List)
			users.POST("", userController.Create)
			users.GET("/:id", userController.GetByID)
			users.PUT("/:id", userController.Update)
			users.DELETE("/:id", userController.Delete)
		}
	}

	return r
}
