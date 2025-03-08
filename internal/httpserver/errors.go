package httpserver

import "errors"

var (
	ErrorTypeConversion     = errors.New("type conversion error")
	ErrorTypeNotImplemented = errors.New("not implemented")
)
