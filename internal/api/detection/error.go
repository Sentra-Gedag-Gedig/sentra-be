package detection

import (
	"ProjectGolang/pkg/response"
	"net/http"
)

var (
	ErrInternalServerError = response.NewError(http.StatusInternalServerError, "internal server error")
	ErrBadRequest          = response.NewError(http.StatusBadRequest, "bad request")
)
