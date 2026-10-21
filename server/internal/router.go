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

	// Health ping endpoints
	pingHandler := func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	}
	r.GET("/ping", pingHandler)
	r.GET("/api/ping", pingHandler)

	// Dependency Injection: Repositories
	userRepo := repositories.NewUserRepository(db)
	credRepo := repositories.NewCredentialRepository(db)
	profileRepo := repositories.NewProfileRepository(db)

	// Dependency Injection: Services
	authService := services.NewAuthService(userRepo, credRepo, profileRepo, c.JWTSecret)
	userService := services.NewUserService(userRepo, credRepo, profileRepo)

	// Dependency Injection: Controllers
	authController := controllers.NewAuthController(authService)
	userController := controllers.NewUserController(userService)

	// Auth Middleware
	authMiddleware := middleware.RequireAuth(c.JWTSecret, credRepo)

	// API Routing: /api/v1
	api := r.Group("/api")
	v1 := api.Group("/v1")
	{
		// Authentication Routes: /api/v1/auth
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authController.Register)
			auth.POST("/login", authController.Login)
			auth.POST("/refresh", authController.RefreshToken)
			auth.GET("/me", authMiddleware, authController.GetMe)
			auth.POST("/logout-all", authMiddleware, authController.LogoutAll)
		}

		// User Management Routes: /api/v1/users
		users := v1.Group("/users")
		{
			users.GET("", userController.List)
			users.POST("", userController.Create)
			users.GET("/:id", userController.GetByID)
			users.PATCH("/:id", userController.Update)
			users.PUT("/:id", userController.Update)
			users.DELETE("/:id", userController.Delete)
		}
	}

	return r
}
