package agent

import "errors"

var (
	ErrorTypeConversion  = errors.New("type conversion error")
	ErrorMarshallingJSON = errors.New("error marshaling json")
)
