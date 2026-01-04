package routes

import (
	"fmt"

	"github.com/dbunt1tled/fiber-go-api/internal/app"
	"github.com/dbunt1tled/fiber-go-api/internal/config"
	"github.com/dbunt1tled/fiber-go-api/internal/lib/view"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/favicon"
)

func WebRoutes(application *app.Application) {
	app := application.Engine()
	app.Get("/", func(c fiber.Ctx) error {
		return c.Render("general/home.gohtml", view.MakeTemplateData(map[string]any{}))
	})
	app.Use(favicon.New(favicon.Config{
		URL:  "/favicon.ico",
		File: fmt.Sprintf("%s/images/favicon.ico", config.Get().Static.Directory),
	}))
}
