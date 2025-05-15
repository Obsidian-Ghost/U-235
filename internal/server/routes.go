package server

import (
	"U-235/handlers"
	"U-235/internal/database"
	"U-235/repositories"
	"U-235/services"
	"U-235/utils"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func (s *Server) RegisterRoutes() http.Handler {

	//Echo Instance
	e := echo.New()

	//Dependencies Initialization
	e.Validator = utils.NewValidator()
	db := database.NewPsqlDB()
	_, _ = database.NewRedisDatabase()
	userRepo := repositories.NewUserRepo(db)
	userService := services.NewUserService(userRepo)
	userHandler := handlers.NewUserHandler(userService)

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

	// Block - User Routes
	{
		//user := e.Group("/user")
		//user.GET("/", handlers.GetUserHandler)
		//user.GET("/urls", handlers.GetUrlsHandler)
	}

	//Block - Core
	{

	}

	//Block - Auth
	{
		auth := api.Group("/auth")
		auth.POST("/register", userHandler.UserRegistrationHandler)
		auth.POST("/login", userHandler.UserLoginHandler)
		//auth.POST("/reset-password",resetPassHandler)
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
