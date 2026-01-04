package http

import (
	"net/http"

	"github.com/dbunt1tled/fiber-go-api/pkg/e"
	"github.com/dbunt1tled/fiber-go-api/pkg/http/dto"
	"github.com/gofiber/fiber/v3"

	"github.com/go-playground/validator/v10"
)

type BaseController struct {
	validation *validator.Validate
}

func NewBaseController(v *validator.Validate) BaseController {
	return BaseController{validation: v}
}

func (b *BaseController) Bind(c fiber.Ctx, dst any, code int) error {
	if err := c.Bind().All(dst); err != nil {
		return e.NewUnprocessableEntityError("invalid body", code)
	}

	return nil
}

func (b *BaseController) BindAndValidate(c fiber.Ctx, dst any, code int) error {
	var err error
	err = b.Bind(c, dst, code)
	if err != nil {
		return err
	}

	if dst, ok := dst.(dto.SetDefaults); ok {
		dst.SetDefaults()
	}

	if err = b.validation.Struct(dst); err != nil {
		return err
	}

	return nil
}

func (b *BaseController) JSON(c fiber.Ctx, status int, data any) error {
	return c.Status(status).JSON(data)
}

func (b *BaseController) JSON200(c fiber.Ctx, data any) error {
	return b.JSON(c, http.StatusOK, data)
}

func (b *BaseController) JSON201(c fiber.Ctx, data any) error {
	return b.JSON(c, http.StatusCreated, data)
}

func (b *BaseController) Error(msg string, code int, status int) error {
	return e.NewErrNo(msg, code, status)
}
