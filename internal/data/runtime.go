package data

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// custom runtime type
type Runtime int32

// error to return if unable to parse or convert the json string
var ErrInvalidRuntimeFormat = errors.New("invalid runtime format")

// custom MarshalJSON that satisfies the json.Marshaler interface
func (r Runtime) MarshalJSON() ([]byte, error) {
	// generate a string containing the runtime in minutes
	jsonValue := fmt.Sprintf("%d mins", r)

	// wrap the string in double quotes to be valid JSON
	quotedJSONValue := strconv.Quote(jsonValue)

	// convert to byte slice and return
	return []byte(quotedJSONValue), nil
}

// custom UnmarshalJSON that satisfies the json.Unmarshaler interface
func (r *Runtime) UnmarshalJSON(jsonValue []byte) error {
	// unquote the json value to remove the double quotes
	unquotedJSONValue, err := strconv.Unquote(string(jsonValue))
	if err != nil {
		return ErrInvalidRuntimeFormat
	}

	parts := strings.Split(unquotedJSONValue, " ")

	if len(parts) != 2 || parts[1] != "mins" {
		return ErrInvalidRuntimeFormat
	}

	// parse the string into an int32
	i, err := strconv.ParseInt(parts[0], 10, 32)
	if err != nil {
		return ErrInvalidRuntimeFormat
	}

	// convert the int32 to a Runtime type and assign to the receiver
	*r = Runtime(i)

	return nil
}
