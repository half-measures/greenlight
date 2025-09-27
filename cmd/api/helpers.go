package main

//HELPERS.GO - For helping!

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

// Get ID url param from request, convert to integer and return it. If bad, return 0 and error
func (app *application) readIDParam(r *http.Request) (int64, error) {
	params := httprouter.ParamsFromContext(r.Context())

	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, errors.New("invalid ID parameter")
	}
	return id, nil
}

//WriteJson helper to send JSON reponses, takes destination http.responsewriter
//http code to send, encodeds to JSON, alters header map if needed

func (app *application) writeJSON(w http.ResponseWriter, status int, data interface{}, headers http.Header) error {
	//encode data to JSON, return error if error
	js, err := json.Marshal(data)
	if err != nil {
		return err
	}
	// append newline to make easy to read in terminal
	js = append(js, '\n')

	// include headers and map to http.responsewriter map
	// Its ok if header is nul, no errors here
	for key, value := range headers {
		w.Header()[key] = value
	}
	// add application/json header and json response
	w.Header().Set("Content-Type", "appilication/json")
	w.WriteHeader(status)
	w.Write(js)
	return nil

}
