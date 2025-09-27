package data

import "time"

//Home to cust movie struct to use all the custom data types for our project, and interacting with the DB itself
//omitempty lets us hide some outputs if they are empty

type Movie struct {
	ID        int64     `json:"id"`                //uniq integer for movie ID
	CreatedAt time.Time `json:"-"`                 //timestamp when movie added to our DB
	Title     string    `json:"title"`             // Movie title
	Year      int32     `json:"year,omitempty"`    //movie release year
	Runtime   int32     `json:"runtime,omitempty"` //in Mins, movie length
	Genres    []string  `json:"genres,omitempty"`  //slice of genres for movie
	Version   int32     `json:"version"`           // Version number, starting at 1 and incredmented ea time movie info updated
}
