package data

import (
	"database/sql"
	"errors"
)

// define cust errnotfound error from our get() method in movies.go
// given when a movie is not in our database
var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict") //Added to prevent data race errors
)

// models struct to wrap moviemodel -
type Models struct {
	Movies      MovieModel
	Permissions PermissionModel //added for avail to handlers and middleware
	Tokens      TokenModel
	Users       UserModel
}

// this method below returns models struct with init movieModel
func NewModels(db *sql.DB) Models {
	return Models{
		Movies:      MovieModel{DB: db},
		Permissions: PermissionModel{DB: db},
		Tokens:      TokenModel{DB: db},
		Users:       UserModel{DB: db},
	} //Done to help later on
}
