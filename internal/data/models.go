package data

import (
	"database/sql"
	"errors"
)

// define cust errnotfound error from our get() method in movies.go
// given when a movie is not in our database
var (
	ErrRecordNotFound = errors.New("record not found")
)

// models struct to wrap moviemodel -
type Models struct {
	Movies MovieModel
}

// this method below returns models struct with init movieModel
func NewModels(db *sql.DB) Models {
	return Models{
		Movies: MovieModel{DB: db},
	} //Done to help later on
}
