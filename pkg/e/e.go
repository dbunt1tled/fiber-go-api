package e

import (
	"fmt"
	"net/http"
)

type ErrNo struct {
	Msg    string `json:"message"`
	Code   int    `json:"code"`
	Status int    `json:"status"`
}

type HTTPError interface {
	Error() string
}

func NewErrNo(msg string, code int, status int) *ErrNo {
	return &ErrNo{
		Msg:    msg,
		Code:   code,
		Status: status,
	}
}

func NewNotFoundError(msg string, code int) HTTPError {
	return NewErrNo(msg, code, http.StatusNotFound)
}

func NewBadRequestError(msg string, code int) HTTPError {
	return NewErrNo(msg, code, http.StatusBadRequest)
}

func NewInternalError(msg string, code int) HTTPError {
	return NewErrNo(msg, code, http.StatusInternalServerError)
}

func NewForbiddenError(msg string, code int) HTTPError {
	return NewErrNo(msg, code, http.StatusForbidden)
}

func NewUnauthorizedError(msg string, code int) HTTPError {
	return NewErrNo(msg, code, http.StatusUnauthorized)
}

func NewValidationError(msg string, code int) HTTPError {
	return NewErrNo(msg, code, http.StatusUnprocessableEntity)
}

func NewUnprocessableEntityError(msg string, code int) HTTPError {
	return NewErrNo(msg, code, http.StatusUnprocessableEntity)
}

func NewUnprocessableEntityErrorWrap(msg string, code int, e error) HTTPError {
	var message string
	if msg == "" {
		message = e.Error()
	} else {
		message = fmt.Sprintf("%s: %s", msg, e.Error())
	}
	return NewErrNo(message, code, http.StatusUnprocessableEntity)
}

func (e ErrNo) Error() string {
	return e.Msg
}
