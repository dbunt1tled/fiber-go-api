package er

import (
	"errors"
	"net/http"

	"github.com/dbunt1tled/fiber-go/pkg/e"
	"github.com/dbunt1tled/fiber-go/pkg/http/dto"
	"github.com/dbunt1tled/fiber-go/pkg/log"
	"github.com/dbunt1tled/fiber-go/pkg/validation"
	"github.com/gofiber/fiber/v3"
)

func APIErrorHandler(ctx fiber.Ctx, err error) error {
	status := http.StatusInternalServerError
	message := err.Error()
	code := 0

	// Handle fiber.Error
	var fiberErr *fiber.Error
	if errors.As(err, &fiberErr) {
		status = fiberErr.Code
	}

	// Handle custom ErrNo
	var errNo *e.ErrNo
	if errors.As(err, &errNo) {
		status = errNo.Status
		message = errNo.Msg // Note: Using Msg instead of Error()
		code = errNo.Code
	}

	vErr := validation.ErrorValidation(err)
	if vErr != nil {
		status = http.StatusUnprocessableEntity
		return ctx.Status(status).JSON(dto.Document{
			Errors: vErr,
		})
	}

	log.Logger().Error(message, err)
	return ctx.Status(status).JSON(dto.Document{
		Errors: []e.ErrNo{{Status: status, Msg: message, Code: code}},
	})
}
