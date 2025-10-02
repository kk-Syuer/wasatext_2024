package service

import "errors"

// ErrNotFound 表示资源不存在
var ErrNotFound = errors.New("not found")

// ErrForbidden 表示权限不足
var ErrForbidden = errors.New("forbidden")

var ErrBadRequest = errors.New("bad request")

var ErrInvalid = errors.New("Invalid request")

