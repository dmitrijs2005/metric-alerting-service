package agent

import "errors"

var (
	ErrorTypeConversion  = errors.New("Type conversion error")
	ErrorMarshallingJson = errors.New("Error marshaling JSON")
)
