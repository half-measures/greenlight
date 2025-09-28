package validator

import "regexp"

//point is to have some small reusable helper functions here

// checking format of email addresses
// This is a basic checker for later stuff. Gives us easy vlaidation checks
var (
	EmailRX = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
)

// new validator type with map of validation errors
type Validator struct {
	Errors map[string]string
}

// new is helper to create new validator instance that is empty
func New() *Validator {
	return &Validator{Errors: make(map[string]string)}
}

// Valid returns true if errors map has no entrys
func (v *Validator) Valid() bool {
	return len(v.Errors) == 0
}

// AddError adds error message to map
func (v *Validator) AddError(key, message string) {
	if _, exists := v.Errors[key]; !exists {
		v.Errors[key] = message
	}
}

// add check error message to map if validation check not ok
func (v *Validator) Check(ok bool, key, message string) {
	if !ok {
		v.AddError(key, message)
	}
}

// Returns true if spesific value is in list of strings
func In(value string, list ...string) bool {
	for i := range list {
		if value == list[i] {
			return true
		}
	}
	return false
}

// Matches true if string value is a regexp pattern
func Matches(value string, rx *regexp.Regexp) bool {
	return rx.MatchString(value)
}

// Uniq returns true if all string values are unique
func Unique(values []string) bool {
	uniqueValues := make(map[string]bool)

	for _, value := range values {
		uniqueValues[value] = true
	}
	return len(values) == len(uniqueValues)
}
