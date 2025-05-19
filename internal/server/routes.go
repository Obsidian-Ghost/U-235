package server

import (
	"U-235/handlers"
	"U-235/internal/database"
	"U-235/repositories"
	"U-235/services"
	"U-235/utils"
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

	// Global API config
	api := e.Group("/api")

	// Block - CORS and Health
	{
		e.Use(middleware.Logger())
		e.Use(middleware.Recover())

		e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
			AllowOrigins:     []string{"https://*", "http://*"},
			AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
			AllowHeaders:     []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
			AllowCredentials: true,
			MaxAge:           300,
		}))

		e.GET("/", s.HelloWorldHandler)

		e.GET("/health", s.healthHandler)

	}

	// Block - URL Management Routes
	{
		urlRoutes := api.Group("/urls")
		urlRoutes.Use(CustomMiddleware.AuthMiddleware)
		urlRoutes.GET("", urlHandler.GetUrlHandler)     // List all URLs for authenticated user
		urlRoutes.POST("", urlHandler.CreateUrlHandler) // Create new shortened URL
		urlRoutes.DELETE("", urlHandler.DeleteUrlHandler)
	}

	// Block - User Profile Routes
	{
		userRoutes := api.Group("/user")
		userRoutes.Use(CustomMiddleware.AuthMiddleware)
		userRoutes.GET("/profile", userHandler.UserProfileHandler) // Get user name, email, and other profile data
		//userRoutes.PUT("/profile", userHandler.UpdateUserProfileHandler) // Update user profile information
	}

	//Block - URL Redirect
	{
		//e.GET("/:shortId", urlHandler.RedirectHandler) // Redirect short URLs to original destination
	}

	//Block - Authentication
	{
		auth := api.Group("/auth")
		auth.POST("/register", userHandler.UserRegistrationHandler)
		auth.POST("/login", userHandler.UserLoginHandler)
		//auth.POST("/forgot-password", userHandler.ForgotPasswordHandler)
	}

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
