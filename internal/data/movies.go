package data

import (
	"database/sql"
	"time"

	"greenlight.alexedwards.net/internal/validator"
)

// Home to cust movie struct to use all the custom data types for our project, and interacting with the DB itself
// omitempty lets us hide some outputs if they are empty
// Note runtime type is custom instead of int32/ If it has 0 value, it will be still empty and ignored
type Movie struct {
	ID        int64     `json:"id"`                //uniq integer for movie ID
	CreatedAt time.Time `json:"-"`                 //timestamp when movie added to our DB
	Title     string    `json:"title"`             // Movie title
	Year      int32     `json:"year,omitempty"`    //movie release year
	Runtime   Runtime   `json:"runtime,omitempty"` //in Mins, movie length
	Genres    []string  `json:"genres,omitempty"`  //slice of genres for movie
	Version   int32     `json:"version"`           // Version number, starting at 1 and incredmented ea time movie info updated
}

// moviemodel struct to wrap a SQL.db connection pool
type MovieModel struct {
	DB *sql.DB
}

// This method will insert a new record into the movies table
func (m MovieModel) Insert(movie *Movie) error {
	return nil //temp
}

// This method will fetch a record from the movies table
func (m MovieModel) Get(id int64) (*Movie, error) {
	return nil, nil
}

// This method will update certian records in movie table
func (m MovieModel) Update(movie *Movie) error {
	return nil
}

// This method will delete a record from Movies table
func (m MovieModel) Delete(id int64) error {
	return nil
}

//Validation checks on the movie STRUCT, not input

func ValidateMovie(v *validator.Validator, movie *Movie) {
	v.Check(movie.Title != "", "title", "must be provided")
	v.Check(len(movie.Title) <= 500, "title", "must not be more than 500 bytes long")

	v.Check(movie.Year != 0, "year", "must be provided")
	v.Check(movie.Year >= 1888, "year", "must be greater than 1888")
	v.Check(movie.Year <= int32(time.Now().Year()), "year", "must not be in the future")

	v.Check(movie.Runtime != 0, "runtime", "must be provided")
	v.Check(movie.Runtime > 0, "runtime", "must be a positive integer")

	v.Check(movie.Genres != nil, "genres", "must be provided")
	v.Check(len(movie.Genres) >= 1, "genres", "must contain at least 1 genre")
	v.Check(len(movie.Genres) <= 5, "genres", "must not contain more than 5 genres")
	v.Check(validator.Unique(movie.Genres), "genres", "must not contain duplicate values")
}
