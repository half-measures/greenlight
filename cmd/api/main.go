package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
	"greenlight.alexedwards.net/internal/data"
)

const version = "1.0.0"

// Declare string for app V number. for now we hardcode

// Config struct for config; used for stuff like network port+
type config struct {
	port int
	env  string
	db   struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  string
	}
}

// app struct to hold HTTP depends, helpers, and middleware.
type application struct {
	config config
	logger *log.Logger
	models data.Models
}

func main() {
	var cfg config //declare config struct

	//read value of port and env cmd line flags in struct
	//default to 4000 and dev if no flags given
	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Enviroment (development|staging|production)")
	flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("GREENLIGHT_DB_DSN"), "PostgreSQL DSN")
	//adding config cmd line limits below
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m", "PostgreSQL max connection idle time")

	flag.Parse()

	//
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	//call openDB function to create connection pool
	//pass in config struct, if err we log it and exit immediately
	db, err := openDB(cfg)
	if err != nil {
		logger.Fatal(err)
	}

	//defer clse so conn pool closed before main func exits
	defer db.Close()

	//log nessage saying pool was success
	logger.Printf("database connection pool established")

	//declare logger struct from app struct
	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
	}

	//create http server with timeouts, using port provided
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	//start the server baby
	logger.Printf("starting %s server on %s", cfg.env, srv.Addr)
	err = srv.ListenAndServe()
	logger.Fatal(err)

}

// returns a connection pool
func openDB(cfg config) (*sql.DB, error) {
	//use sql.open to create empty conn pool
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}
	//max num of idle + inuse sql conns in the pool
	//setting to less than or equal to 0 means unlimited
	db.SetMaxOpenConns(cfg.db.maxOpenConns)
	//Max num of idle connections in pool, same login on 0 as above
	db.SetMaxIdleConns(cfg.db.maxIdleConns)

	//use time to convert idle timeout duration
	duration, err := time.ParseDuration(cfg.db.maxIdleTime)
	if err != nil {
		return nil, err
	}
	//set max idle timeout
	db.SetConnMaxIdleTime(duration)

	//create context wiht 5s timeout Deadline
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	//use ping to create new conn to DB, passing in above as parameter
	//if no conn in 5s, return err
	err = db.PingContext(ctx)
	if err != nil {
		return nil, err

	}
	//return db conn pool
	return db, nil
}
