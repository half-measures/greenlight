package data

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// To do a custom MarshalJSON method on our custom type.
// Putting it here to avoid cluttering movies.go
// Define err unmarshalJSON will return if unable to parse json
var ErrInvalidRuntimeFormat = errors.New("invalid runtime format")

type Runtime int32

//Do marshalJSON method on runtime to satisfies internal
//json.marshal interface. Should return JSON encode value for movie runtime

func (r Runtime) MarshalJSON() ([]byte, error) {
	//gen a string with movie runtime in needed format
	jsonValue := fmt.Sprintf("%d mins", r)

	//use strconv func on string to wrap it in double quotes,
	//needed to be a 'valid' JSON string
	quotedJSONValue := strconv.Quote(jsonValue)
	//convert to byte slice and return it
	return []byte(quotedJSONValue), nil
}

func (r *Runtime) UnmarshalJSON(jsonValue []byte) error {
	//we  expect JSON value will be string
	//need to remove surrounding double quotes from sring
	//if unable, return err
	unquotedJSONValue, err := strconv.Unquote(string(jsonValue))
	if err != nil {
		return ErrInvalidRuntimeFormat
	}

	//split the string to isolate number
	parts := strings.Split(unquotedJSONValue, " ")

	//check to make sure its in expected format
	if len(parts) != 2 || parts[1] != "mins" {
		return ErrInvalidRuntimeFormat
	}
	//Otherwise, parse string with number into int32 again,
	i, err := strconv.ParseInt(parts[0], 10, 32)
	if err != nil {
		return ErrInvalidRuntimeFormat
	}

	//convert int32 type to runtime special type
	//using a pointer to the runtime type to set value of pointer
	*r = Runtime(i)
	return nil
}
