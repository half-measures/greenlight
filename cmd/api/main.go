package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

const version = "1.0.0"

// Declare string for app V number. for now we hardcode

// Config struct for config; used for stuff like network port+
type config struct {
	port int
	env  string
}

// app struct to hold HTTP depends, helpers, and middleware.
type application struct {
	config config
	logger *log.Logger
}

func main() {
	var cfg config //declare config struct

	//read value of port and env cmd line flags in struct
	//default to 4000 and dev if no flags given
	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Enviroment (development|staging|production)")
	flag.Parse()

	//
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	//declare logger struct from app struct
	app := &application{
		config: cfg,
		logger: logger,
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
	err := srv.ListenAndServe()
	logger.Fatal(err)

}
