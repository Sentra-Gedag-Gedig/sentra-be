package auth

import (
	"ProjectGolang/pkg/response"
	"net/http"
)

var (
	ErrPhoneNumberAlreadyExists = response.NewError(http.StatusConflict, "phone number already exists")
	ErrEmailAlreadyExists       = response.NewError(http.StatusConflict, "email already exists")
	ErrInvalidEmailOrPassword   = response.NewError(http.StatusBadRequest, "email or password is wrong")
	ErrUserNotFound             = response.NewError(http.StatusNotFound, "user not found")
	ErrorInvalidToken           = response.NewError(http.StatusUnauthorized, "invalid token")
	ErrUserWithEmailNotFound    = response.NewError(http.StatusNotFound, "user with email not found")
	ErrPasswordSame             = response.NewError(http.StatusBadRequest, "password same as before")
	ErrorTokenExpired           = response.NewError(http.StatusBadRequest, "token expired or not found")
	ErrInvalidPhoneNumber       = response.NewError(http.StatusBadRequest, "invalid phone number")
	ErrInvalidOTP               = response.NewError(http.StatusBadRequest, "invalid otp")
	ErrInvalidToken             = response.NewError(http.StatusBadRequest, "invalid token")
	ErrInvalidFileType          = response.NewError(http.StatusBadRequest, "invalid file type")
	ErrFileTooLarge             = response.NewError(http.StatusBadRequest, "file too large")
	ErrFailedToUploadFile       = response.NewError(http.StatusInternalServerError, "failed to upload file")
	ErrInvalidEmail             = response.NewError(http.StatusBadRequest, "invalid email")
	ErrEmailAlreadyInUse        = response.NewError(http.StatusConflict, "email already in use by another user")
)
