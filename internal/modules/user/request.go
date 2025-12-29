package user

import "github.com/dbunt1tled/fiber-go/pkg/http/dto"

type ListRequest struct {
	dto.PaginationQuery

	Email  *string   `query:"email"  json:"email"    validate:"omitempty,email,max=70" example:"fake@example.com"`
	Status *[]Status `query:"status" json:"password" validate:"omitempty"              example:"0,1"`
	Roles  *Roles    `query:"roles"  json:"roles"    validate:"omitempty"              example:"admin,person"`
}
