package services

import "errors"

var (
	ErrConflict           = errors.New("conflict")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrForbidden          = errors.New("forbidden")
	ErrNotFound           = errors.New("not found")
	ErrBadRequest         = errors.New("bad request")
	ErrRateLimited        = errors.New("rate limited")
)
