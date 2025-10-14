package main

import (
	"context"
	"database/sql"
	"flag"
	"os"
	"strings"
	"sync"
	"time"

	_ "github.com/lib/pq"
	"greenlight.alexedwards.net/internal/data"
	"greenlight.alexedwards.net/internal/jsonlog"
	"greenlight.alexedwards.net/internal/mailer"
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
	// new struct for req per second, can enable or disable Rate limits also
	limiter struct {
		rps     float64
		burst   int
		enabled bool
	}
	smtp struct { //holds email info
		host     string
		port     int
		username string
		password string
		sender   string
	}
	//for middleware.go - prevent CORS
	cors struct {
		trustedOrigins []string
	}
}

// app struct to hold HTTP depends, helpers, and middleware.
// Note the custom logger call here
type application struct {
	config config
	logger *jsonlog.Logger
	models data.Models
	mailer mailer.Mailer
	wg     sync.WaitGroup //Used to allow very graceful shutdown with sync.waitgroups
}

func main() {
	var cfg config         //declare config struct
	secrets := getSecret() //from secrets.go and our config file

	//read value of port and env cmd line flags in struct
	//default to 4000 and dev if no flags given
	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Enviroment (development|staging|production)")
	flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("GREENLIGHT_DB_DSN"), "PostgreSQL DSN")
	//adding config cmd line limits below
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m", "PostgreSQL max connection idle time")

	//Below is the Cmd line settings to set values in config struct
	//default is enabled rate limiting
	flag.Float64Var(&cfg.limiter.rps, "limiter-rps", 2, "Rate limiter maximum requests per second")
	flag.IntVar(&cfg.limiter.burst, "limiter-burst", 4, "Rate limiter maximum burst")
	flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", true, "enabled Rate limiter")

	username := secrets.MailtrapUsername //getting returned fields
	password := secrets.MailtrapPassword
	//read SMTP server config settings into config struct
	flag.StringVar(&cfg.smtp.host, "smtp-host", "smtp.mailtrap.io", "SMTP host")
	flag.IntVar(&cfg.smtp.port, "smtp-port", 587, "SMTP port") //port 25 failed for some reason
	flag.StringVar(&cfg.smtp.username, "smtp-username", username, "SMTP username")
	flag.StringVar(&cfg.smtp.password, "smtp-password", password, "SMTP password")
	flag.StringVar(&cfg.smtp.sender, "smtp-sender", "Greenlight <no-reply@greenlight.alexedwards.net>", "SMTP sender")

	//Flag.Func to process cors trusted origins CLI flag at runtime
	//we split it and assign it to our config struct
	flag.Func("cors-trusted-origins", "Trusted CORS origins (space separated)", func(val string) error {
		cfg.cors.trustedOrigins = strings.Fields(val)
		return nil
	})

	flag.Parse()

	//init cust logger for any err at or above INFO to outscreen
	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)

	//call openDB function to create connection pool
	//pass in config struct, if err we log it and exit immediately
	db, err := openDB(cfg)
	if err != nil {
		logger.PrintFatal(err, nil)
	}

	//defer clse so conn pool closed before main func exits
	defer db.Close()

	//log nessage saying pool was success
	logger.PrintInfo("database connection pool established", nil)

	//declare logger struct from app struct
	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
		mailer: mailer.New(cfg.smtp.host, cfg.smtp.port, cfg.smtp.username, cfg.smtp.password, cfg.smtp.sender),
	} //Mailer instance into application struct

	//create http server with timeouts, using port provided - moved to server.go
	err = app.serve()
	if err != nil {
		logger.PrintFatal(err, nil)
	}

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
