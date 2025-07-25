package common

import (
	"errors"
	"fmt"
)

var (
	ErrorTypeConversion      = errors.New("type conversion error")
	ErrorMarshallingJSON     = errors.New("error marshaling json")
	ErrorMetricDoesNotExist  = errors.New("metric does not exist")
	ErrorMetricAlreadyExists = errors.New("metric already exists")
	ErrorTypeNotImplemented  = errors.New("not implemented")
)

type WrappedError struct {
	Label string
	Err   error
}

func NewWrappedError(label string, err error) *WrappedError {
	return &WrappedError{Label: label, Err: err}
}

func (e *WrappedError) Error() string {
	return fmt.Sprintf("%s %v", e.Label, e.Err)
}

func (e *WrappedError) Unwrap() error {
	return e.Err
}
