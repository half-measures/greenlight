package data

//used to change and talk to our DB
import (
	"database/sql"
	"errors"
	"time"

	"github.com/lib/pq"
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
//accepts a pointer to movie struct, which should have data for new record

func (m MovieModel) Insert(movie *Movie) error {
	// define SQL query for new record
	query := `
	INSERT INTO movies (title, year, runtime, genres)
	VALUES ($1, $2, $3, $4)
	RETURNING id, created_at, version`

	//create slice with values , doing this next to the SQL query
	//makes it clear
	args := []interface{}{movie.Title, movie.Year, movie.Runtime, pq.Array(movie.Genres)}

	//use QueryRow to exec SQL on connection pool
	//string gets passes in as variadic parameter
	return m.DB.QueryRow(query, args...).Scan(&movie.ID, &movie.CreatedAt, &movie.Version)
} //Insert mutates moviestruct and adds system gen values to it

// This method will fetch a record from the movies table
func (m MovieModel) Get(id int64) (*Movie, error) {
	//Define SQL query for GET
	if id < 1 { //This is to align ourselves with postgres as it dosen't have unsigned integers
		//and to prevent a value more than 92233720365457758....
		return nil, ErrRecordNotFound
	}

	query := `
	SELECT id, created_at, title, year, runtime, genres, version
	FROM movies
	WHERE id = $1`
	//declare Movie struct to hold the movie data
	var movie Movie

	//Execute using queryrow, scan response data into fields into
	//movie struct, use pq.array adapter function
	err := m.DB.QueryRow(query, id).Scan(
		&movie.ID,
		&movie.CreatedAt,
		&movie.Title,
		&movie.Year,
		&movie.Runtime,
		pq.Array(&movie.Genres),
		&movie.Version,
	) //if we did not use pq.Array would get an error at runtime
	//'unsupported Scan...

	//Handle Errors, if no match found scan return sql.errnorows
	//errs, check for this
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	//otherwise return pointer to movie struct
	return &movie, nil
}

// This method will update certian records in movie table
func (m MovieModel) Update(movie *Movie) error {
	//SQL to update record
	query := `
	UPDATE movies
	SET title = $1, year = $2, runtime = $3, genres = $4, version = version + 1
	WHERE id = $5 AND version = $6
	RETURNING version`

	//args slice to hold values of placeholder params we overwrite later
	args := []interface{}{
		movie.Title,
		movie.Year,
		movie.Runtime,
		pq.Array(movie.Genres),
		movie.ID,
		movie.Version,
	}
	//queryrow to execute query on arge slice, scan new version into movie struct
	err := m.DB.QueryRow(query, args...).Scan(&movie.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}
	return nil
	//mutate in place again, update wiht new version num only
}

// This method will delete a record from Movies table
func (m MovieModel) Delete(id int64) error {
	//check, return errrecordnotfound if movie if less than 1
	if id < 1 {
		return ErrRecordNotFound
	}
	//SQL
	query := `
	DELETE FROM movies
	WHERE id = $1`

	//exec SQL query
	result, err := m.DB.Exec(query, id)
	if err != nil {
		return err
	}

	//call rows affected method on result object
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	// If no rows affected, table did not have record
	//provide errrecordnotfound error
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}
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
