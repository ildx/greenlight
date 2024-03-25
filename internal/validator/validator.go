package validator

import (
	"regexp"
	"slices"
)

// whatwg email regex
// https://html.spec.whatwg.org/#valid-e-mail-address
var (
	EmailRX = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
)

// validator type
type Validator struct {
	Errors map[string]string
}

// New is a helper which creates a new Validator instance with an empty errors map
func New() *Validator {
	return &Validator{Errors: make(map[string]string)}
}

// Valid returns true if no entries in error map
func (v *Validator) Valid() bool {
	return len(v.Errors) == 0
}

// AddError adds an error message to the map (so long as the message is not empty)
func (v *Validator) AddError(key, message string) {
	if _, exists := v.Errors[key]; !exists {
		v.Errors[key] = message
	}
}

// Check adds an error message to the map only if a validation check is not ok
func (v *Validator) Check(ok bool, key, message string) {
	if !ok {
		v.AddError(key, message)
	}
}

// PermittedValue checks if a value is in a list of permitted values
func PermittedValue[T comparable](value T, permittedValues ...T) bool {
	return slices.Contains(permittedValues, value)
}

// Matches return true if a string value matches a specific regex pattern
func Matches(value string, rx *regexp.Regexp) bool {
  return rx.MatchString(value)
}

// Generic function to check if all values in a slice are unique
func Unique[T comparable](values []T) bool {
  uniqueValues := make(map[T]bool)

  for _, value := range values {
    uniqueValues[value] = true
  }

  return len(values) == len(uniqueValues)
}
