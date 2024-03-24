package data

import (
	"fmt"
	"strconv"
)

// custom runtime type
type Runtime int32

// custom MarshalJSON that satisfies the json.Marshaler interface
func (r Runtime) MarshalJSON() ([]byte, error) {
	// generate a string containing the runtime in minutes
	jsonValue := fmt.Sprintf("%d mins", r)

	// wrap the string in double quotes to be valid JSON
	quotedJSONValue := strconv.Quote(jsonValue)

	// convert to byte slice and return
	return []byte(quotedJSONValue), nil
}
