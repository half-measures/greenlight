package main

//HELPERS.GO - For helping!

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"
	"greenlight.alexedwards.net/internal/validator"
)

// Define envelope Type to add custom envelope map to map[string]interface{}
type envelope map[string]interface{}

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

func (app *application) writeJSON(w http.ResponseWriter, status int, data envelope, headers http.Header) error {
	//encode data to JSON, return error if error
	//using no line prefix "" and tab indents \t for each element returned
	js, err := json.MarshalIndent(data, "", "\t")
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

// Helps decode JSON from request body as normal. Doing this to avoid
// letting our public API give too much info away about how it works
func (app *application) readJSON(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	maxBytes := 1_048_576 //limit size of req to 1mb
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	//init decoder and do disallow unknown fields on it.
	//Now if decoder gets a unknown field it will error instead of ignoring it
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	//decode reqest to target
	err := dec.Decode(dst)
	if err != nil {
		//Triage the errors cause and output our own stuff instead
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError

		switch {
		//use error.as to check if error has type
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formed JSON at char %d", syntaxError.Offset)

		//errunexpectedEOF err, check
		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")

		//catch any unmarshaltype errors - when JSON value is wrong type for destination
		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains bad JSON type for respective field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type at %d", unmarshalTypeError.Offset)

		//check for EOF error if request is empty
		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")

		//if non-nil pointer gotten, we panic rather than returning error
		case errors.As(err, &invalidUnmarshalError):
			panic(err)
		//for all else, return err as is
		//Panic be special here, shoulden't be seen under normal ops
		default:
			return err
		}
	}
	//call decode using pointer to anon struct as target dest.
	//If request has only a single JSON value, its a EOF err, if we get anything
	//else, theres addtional data
	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must contains single JSON value")
	}
	return nil
}

// ReadString returns string value from query string
// returns default value if no matching key found
func (app *application) readString(qs url.Values, key string, defaultValue string) string {
	//extract value key from query string. If no key exists return empty ""

	s := qs.Get(key)
	//if no key exists, or alue empty, return default value
	if s == "" {
		return defaultValue
	}
	//return string otherwise
	return s
}

// readCSV helper reads string value form query string, splits it
// into slice of comma characters. If no match, return default value
func (app *application) readCSV(qs url.Values, key string, defaultValue []string) []string {
	//extract value from query string
	csv := qs.Get(key)
	//if no key exists or is empty, return default value
	if csv == "" {
		return defaultValue
	}
	//otherwise parse value into []string slice and return it
	return strings.Split(csv, ",")
}

// readint helper reads string value from query string and converts it to integer
// returns it then. If no match match return default value
// If could not convert to integer, record err,
func (app *application) readInt(qs url.Values, key string, defaultValue int, v *validator.Validator) int {
	//extract value from query string
	s := qs.Get(key)
	//if no key exists or is empty, return default value
	if s == "" {
		return defaultValue
	}
	//try to convert value to a int. If failes, err message to validator
	i, err := strconv.Atoi(s)
	if err != nil {
		v.AddError(key, "must be a integer value")
		return defaultValue
	}
	//otherwise, return correctly converted integer value
	return i
}

// help function to try  and wrap the recovering logic
// uses Go's first class functions, where functions can be assigned to variables
// and passed as parameters to other functions
// background accepts any function wiht func() as param, stores it in fn var
// its in the backgroun to try and recover any panic and logs err,
func (app *application) background(fn func()) {
	// Increment the WaitGroup counter.
	app.wg.Add(1)

	// Launch the background goroutine.
	go func() {
		// Use defer to decrement the WaitGroup counter before the goroutine returns.
		defer app.wg.Done()

		defer func() {
			if err := recover(); err != nil {
				app.logger.PrintError(fmt.Errorf("%s", err), nil)
			}
		}()

		fn()
	}()
}
