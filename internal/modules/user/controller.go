package user

import (
	"github.com/dbunt1tled/fiber-go/pkg/e"
	"github.com/dbunt1tled/fiber-go/pkg/http"
	"github.com/dbunt1tled/fiber-go/pkg/storage"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
)

type Controller struct {
	http.BaseController

	userService *Service
}

func NewUserController(
	userService *Service,
	validation *validator.Validate,
) *Controller {
	return &Controller{
		BaseController: http.NewBaseController(validation),
		userService:    userService,
	}
}

func (uc *Controller) List(c fiber.Ctx) error {
	var (
		err   error
		users *storage.Paginator[*User]
	)
	req := new(ListRequest)

	if err = uc.BindAndValidate(c, req, e.Err422UserListValidateError); err != nil {
		return err
	}

	users, err = uc.userService.Paginate(
		c.Context(),
		req.Page.Page,
		req.Page.Limit,
		storage.WithFilter(
			storage.NewRule("status", storage.OpIn, req.Status),
			storage.NewRule("email", storage.OpEqual, req.Email),
			storage.NewRule("roles", storage.OpContains, req.Roles),
		),
		storage.WithSort(req.Sort.Field, req.Sort.Order),
	)

	if err != nil {
		return e.NewUnprocessableEntityError(
			err.Error(),
			e.Err422UserListError,
		)
	}

	return uc.JSON200(c, NewUserListResponse(users))
}
