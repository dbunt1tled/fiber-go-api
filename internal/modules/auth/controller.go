package auth

import (
	"time"

	"github.com/dbunt1tled/fiber-go-api/internal/lib/aemail"
	"github.com/dbunt1tled/fiber-go-api/internal/modules/user"
	"github.com/dbunt1tled/fiber-go-api/pkg/e"
	"github.com/dbunt1tled/fiber-go-api/pkg/f"
	"github.com/dbunt1tled/fiber-go-api/pkg/hasher"
	"github.com/dbunt1tled/fiber-go-api/pkg/http"
	"github.com/dbunt1tled/fiber-go-api/pkg/storage"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type Controller struct {
	http.BaseController

	authService *Service
	userService *user.Service
	mailService *aemail.MailServiceAsync
}

func NewController(
	authService *Service,
	userService *user.Service,
	mailService *aemail.MailServiceAsync,
	validation *validator.Validate,
) *Controller {
	return &Controller{
		BaseController: http.NewBaseController(validation),
		authService:    authService,
		userService:    userService,
		mailService:    mailService,
	}
}

func (a *Controller) Register(c fiber.Ctx) error {
	var (
		err                    error
		password, confirmToken string
		u                      *user.User
	)
	req := new(Register)

	if err = a.BindAndValidate(c, req, e.Err422RegisterValidateError); err != nil {
		return err
	}

	password, err = a.authService.GeneratePasswordHash(req.Password)
	if err != nil {
		return e.NewUnprocessableEntityErrorWrap(
			"Password hash error.",
			e.Err422RegisterUserPasswordError,
			err,
		)
	}
	u, err = a.userService.Create(c.Context(), req.ToUser().WithPassword(password))
	if err != nil {
		return e.NewUnprocessableEntityErrorWrap(
			"User creation error.",
			e.Err422RegisterUserCreationError,
			err,
		)
	}

	confirmToken, err = a.authService.GenerateConfirmToken(u)
	if err != nil {
		return e.NewUnprocessableEntityErrorWrap(
			"Generate token error.",
			e.Err422CreateConfirmTokenError,
			err,
		)
	}

	err = a.mailService.SendConfirmMail(u, confirmToken)
	if err != nil {
		return e.NewUnprocessableEntityErrorWrap(
			"Send email error.",
			e.Err422SendConfirmEmailError,
			err,
		)
	}

	return a.JSON200(c, user.NewUserResponse(u))
}

func (a *Controller) Confirm(c fiber.Ctx) error {
	var (
		id    uuid.UUID
		err   error
		token map[string]interface{}
		u     *user.User
	)
	req := new(Confirm)

	if err = a.BindAndValidate(c, req, e.Err422ConfirmValidateError); err != nil {
		return err
	}

	token, err = a.authService.DecodeConfirmToken(req.Token)

	if err != nil {
		return e.NewUnprocessableEntityErrorWrap(
			"confirm user error",
			e.Err422ConfirmTokenError,
			err,
		)
	}
	id, err = uuid.Parse(token["iss"].(string))
	if err != nil {
		return e.NewUnprocessableEntityError("confirm user error", e.Err422TokenConfirmUserIdError)
	}

	u, err = a.userService.FindByID(c.Context(), id)
	if err != nil {
		return e.NewUnprocessableEntityError("confirm user error", e.Err422TokenConfirmUserError)
	}

	if u == nil {
		return e.NewUnprocessableEntityError("confirm user error", e.Err422TokenConfirmUserNotFoundError)
	}

	if u.Status != user.Pending {
		return e.NewUnprocessableEntityError("confirm user error", e.Err422TokenConfirmUserNotPendingError)
	}
	u.Status = user.Active
	u.ConfirmedAt = f.Pointer(time.Now())
	u, err = a.userService.Update(c.Context(), u)

	if err != nil {
		return e.NewUnprocessableEntityError("confirm user error", e.Err422TokenConfirmUserUpdateError)
	}

	return a.JSON200(c, user.NewUserResponse(u))
}

func (a *Controller) Login(c fiber.Ctx) error {
	var (
		err error
		u   *user.User
		ch  bool
	)
	req := new(Login)

	if err = a.BindAndValidate(c, req, e.Err422LoginValidateError); err != nil {
		return err
	}

	u, err = a.userService.One(
		c.Context(),
		storage.WithFilter(storage.NewRule("status", storage.OpEqual, user.Active)),
	)
	if err != nil {
		return e.NewUnprocessableEntityError(
			"Authorization error, password or login is incorrect.",
			e.Err422LoginUserNotFoundError,
		)
	}

	ch, err = a.authService.ValidatePassword(req.Password, u.Password)
	if err != nil {
		return e.NewUnprocessableEntityError(
			err.Error(),
			e.Err422LoginUserPasswordError,
		)
	}

	if !ch {
		return e.NewUnprocessableEntityError(
			"Authorization error, password or login is incorrect.",
			e.Err422LoginUserPasswordWrongError,
		)
	}

	access, refresh, err := a.authService.GenerateAuthTokens(u)
	if err != nil {
		return e.NewUnprocessableEntityError(
			err.Error(),
			e.Err422LoginAccessTokenError,
		)
	}

	return a.JSON200(c, NewLoginResponse(map[string]interface{}{
		"accessToken":  access,
		"refreshToken": refresh,
	}))
}

func (a *Controller) Refresh(c fiber.Ctx) error {
	auth := c.Get("Authorization")
	if auth == "" {
		return e.NewUnauthorizedError("Unauthorized", e.Err401RefreshEmptyTokenError)
	}

	token, err := a.authService.DecodeBearerToken(auth, hasher.WithSubject(hasher.RefreshTokenSubject))
	if err != nil {
		return err
	}

	id, err := uuid.Parse(token["iss"].(string))
	if err != nil {
		return e.NewUnauthorizedError("Unauthorized", e.Err401TokenRefreshUserIdError)
	}

	u, err := a.userService.FindByID(c.Context(), id)
	if err != nil {
		return e.NewUnauthorizedError("Unauthorized", e.Err401TokenRefreshUserError)
	}

	if u == nil {
		return e.NewUnauthorizedError("Unauthorized", e.Err401RefreshUserNotFoundError)
	}

	if u.Status != user.Active {
		return e.NewUnauthorizedError("Unauthorized", e.Err401RefreshUserNotActiveError)
	}

	access, refresh, err := a.authService.GenerateAuthTokens(u)
	if err != nil {
		return e.NewUnprocessableEntityError(
			err.Error(),
			e.Err422LoginRefreshTokenError,
		)
	}

	return a.JSON200(c, NewLoginResponse(map[string]interface{}{
		"accessToken":  access,
		"refreshToken": refresh,
	}))
}
