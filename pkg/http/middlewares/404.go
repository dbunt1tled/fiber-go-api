package middlewares

import (
	"github.com/dbunt1tled/fiber-go/pkg/e"
	"github.com/gofiber/fiber/v3"
)

func NotFound(c fiber.Ctx) error {
	return e.NewNotFoundError("404 Not Found", e.Err404NotFoundDefault)
}
