package middlewares

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/dbunt1tled/fiber-go/pkg/e"
	"github.com/dbunt1tled/fiber-go/pkg/log"
	"github.com/gofiber/fiber/v3"
)

func NewLog(skip func(c fiber.Ctx) bool) fiber.Handler {
	return func(c fiber.Ctx) error {
		start := time.Now()
		next := c.Next()

		if skip != nil && skip(c) {
			return nil
		}
		status := c.Response().StatusCode()

		msg := "ОK"
		if next != nil {
			status = http.StatusInternalServerError
			msg = next.Error()
			var errNo *e.ErrNo
			if errors.As(next, &errNo) {
				status = errNo.Status
			}
		}

		log.Logger().Log(
			c.Context(),
			parseLevel(status),
			fmt.Sprintf("[HTTP] request %s (%s) %s", msg, c.Method(), c.FullPath()),
			slog.String("url", c.OriginalURL()),
			slog.Any("headers", c.GetHeaders()),
			slog.String("remoteAddr", c.IP()),
			slog.String("reqBody", string(c.Request().Body())),
			slog.String("userAgent", c.Get(fiber.HeaderUserAgent)),
			slog.Int("status", status),
			slog.String("duration", time.Since(start).Round(time.Millisecond).String()),
			slog.String("resBody", string(c.Response().Body())),
		)

		return next
	}
}

func parseLevel(status int) slog.Level {
	switch {
	case status < http.StatusMultipleChoices:
		return slog.LevelDebug
	case status >= http.StatusMultipleChoices && status < http.StatusBadRequest:
		return slog.LevelInfo
	case status >= http.StatusBadRequest && status < http.StatusInternalServerError:
		return slog.LevelWarn
	case status >= http.StatusInternalServerError:
		return slog.LevelError
	default:
		return slog.LevelDebug
	}
}
