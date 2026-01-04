package middlewares

import (
	"github.com/dbunt1tled/fiber-go-api/internal/modules/auth"
	"github.com/dbunt1tled/fiber-go-api/internal/modules/user"
	"github.com/dbunt1tled/fiber-go-api/pkg/e"
	"github.com/dbunt1tled/fiber-go-api/pkg/hasher"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type AuthMiddleware struct {
	authService *auth.Service
	userService *user.Service
}

func NewAuthMiddleware(
	authService *auth.Service,
	userService *user.Service,
) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
		userService: userService,
	}
}

func (a *AuthMiddleware) Auth(c fiber.Ctx) error {
	authorization := c.Get("Authorization")
	if authorization == "" {
		return e.NewUnauthorizedError("Unauthorized", e.Err401AuthEmptyTokenError)
	}

	token, err := a.authService.DecodeBearerToken(authorization, hasher.WithSubject(hasher.AccessTokenSubject))
	if err != nil {
		return err
	}

	id, err := uuid.Parse(token["iss"].(string))
	if err != nil {
		return e.NewUnauthorizedError("Unauthorized", e.Err401TokenUserIdError)
	}

	u, err := a.userService.FindByID(c.Context(), id)
	if err != nil {
		return e.NewUnauthorizedError("Unauthorized", e.Err401TokenUserIdError)
	}

	if u == nil {
		return e.NewUnauthorizedError("Unauthorized", e.Err401UserNotFoundError)
	}

	if u.Status != user.Active {
		return e.NewUnauthorizedError("Unauthorized", e.Err401UserNotActiveError)
	}

	c.Locals("user", u)

	return c.Next()
}
