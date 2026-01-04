package routes

import (
	"fmt"

	"github.com/dbunt1tled/fiber-go-api/internal/app"
	"github.com/dbunt1tled/fiber-go-api/internal/config"
	"github.com/gofiber/fiber/v3"
)

func ApiRoutes(application *app.Application) {
	app := application.Engine()
	api := app.Group("api")
	apiRoutes(api, application)
}

func apiRoutes(api fiber.Router, a *app.Application) {
	api.Get("/", a.AuthMiddleware.Auth, func(c fiber.Ctx) error {
		return c.SendString(fmt.Sprintf("%s(%s)", config.Get().Name, config.Get().Env))
	})
	authGroup := api.Group("auth")
	apiAuthRoutes(authGroup, a)

	userGroup := api.Group("users", a.AuthMiddleware.Auth)
	apiUserRoutes(userGroup, a)
}

func apiUserRoutes(userGroup fiber.Router, a *app.Application) {
	userGroup.Get("/", a.UserController.List)
}

func apiAuthRoutes(authGroup fiber.Router, a *app.Application) {
	authGroup.Post("/login", a.AuthController.Login)
	authGroup.Post("/refresh", a.AuthController.Refresh)
	authGroup.Post("/register", a.AuthController.Register)
	authGroup.Get("/confirm/:token", a.AuthController.Confirm)
}
