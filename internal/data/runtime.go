package data

import (
	"fmt"
	"strconv"
)

//To do a custom MarshalJSON method on our custom type.
//Putting it here to avoid cluttering movies.go

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
