package server

import (
	"U-235/handlers"
	"U-235/internal/database"
	"U-235/repositories"
	"U-235/services"
	"U-235/utils"
	"context"
	"log"
	"net/http"

	CustomMiddleware "U-235/middleware"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func (s *Server) RegisterRoutes() http.Handler {

	//Echo Instance
	e := echo.New()

	//Dependencies Initialization
	e.Validator = utils.NewValidator()
	db := database.NewPsqlDB()
	gormDB := database.NewGormPostgresDB()
	redisDB, _ := database.NewRedisDatabase()
	userRepo := repositories.NewUserRepo(db)
	userService := services.NewUserService(userRepo)
	userHandler := handlers.NewUserHandler(userService)

	psqlRepo := repositories.NewUrlsPsql(db, gormDB)
	redisRepo, _ := repositories.NewUrlRedis(redisDB)
	urlService := services.NewShortUrlService(redisRepo, psqlRepo)
	urlHandler := handlers.NewUrlHandler(urlService)

	// Add expiration service initialization
	expirationService := services.NewRedisExpirationService(redisDB, psqlRepo)
	ctx := context.Background()

	// Initialize and start expiration handler
	if err := expirationService.InitializeKeyspaceNotifications(ctx); err != nil {
		log.Printf("Warning: Failed to initialize keyspace notifications: %v", err)
	} else {
		expirationService.StartExpirationListener(ctx)
	}

	// Global middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"https://*", "http://*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// IMPORTANT: Register specific routes FIRST before the catch-all route

	// Health and system routes
	e.GET("/", s.HelloWorldHandler)
	e.GET("/health", s.healthHandler)

	// Global API config - All API routes under /api
	api := e.Group("/api")

	// Authentication routes
	{
		auth := api.Group("/auth")
		auth.POST("/register", userHandler.UserRegistrationHandler)
		auth.POST("/login", userHandler.UserLoginHandler)
		// auth.POST("/forgot-password", userHandler.ForgotPasswordHandler)
	}

	// URL Management Routes (authenticated)
	{
		urlRoutes := api.Group("/urls")
		urlRoutes.Use(CustomMiddleware.AuthMiddleware)
		urlRoutes.GET("", urlHandler.GetUrlHandler)
		urlRoutes.POST("", urlHandler.CreateUrlHandler)
		urlRoutes.DELETE("/:urlId", urlHandler.DeleteUrlHandler)
		urlRoutes.POST("/expiry", urlHandler.ExtendExpiryHandler)
	}

	// User Profile Routes (authenticated)
	{
		userRoutes := api.Group("/user")
		userRoutes.Use(CustomMiddleware.AuthMiddleware)
		userRoutes.GET("/profile", userHandler.UserProfileHandler)
	}

	// CRITICAL: Register the catch-all route LAST
	// This ensures all specific routes are matched first
	api.GET("/redirect/:shortId", urlHandler.RedirectHandler,
		CustomMiddleware.ValidateShortId,
		CustomMiddleware.UrlCache,
	)

	return e
}

func (s *Server) HelloWorldHandler(c echo.Context) error {
	resp := map[string]string{
		"message": "Hello World",
	}

	return c.JSON(http.StatusOK, resp)
}

func (s *Server) healthHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, s.db.Health())
}
