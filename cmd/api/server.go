package main

//meant to allow graceful start and stop of the server.
import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func (app *application) serve() error {
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.config.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	//shutdown channel to get any errs
	shutdownError := make(chan error)

	//background roroutine for catching sig errors to allow grace shutdown
	//This is a buffered channel with size 1 as signalNotify does not wait for receiver to be avail
	//we dont want the signal missed
	go func() {
		//create quit channel to carry os.Signal value
		quit := make(chan os.Signal, 1)
		//use notify to listen for any SIGINT & SIGTERM and relay them
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		//read signal
		s := <-quit
		//log message to say its been caught
		app.logger.PrintInfo("caught signal", map[string]string{
			"signal": s.String(),
		})
		//context with 5s timeout
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		//call shutdown on serverm passin in ctx
		//will return nil if gracefulshutdown was good or err
		err := srv.Shutdown(ctx)
		if err != nil {
			shutdownError <- err
		}
		//log message that were waiting for background goroutines to complete
		app.logger.PrintInfo("Completing background tasks", map[string]string{
			"addr": srv.Addr,
		})
		//Call wait to block till waitGroup counter is zero
		//block till routines have completes
		app.wg.Wait()
		shutdownError <- nil
	}()
	//display starting server message
	app.logger.PrintInfo("starting server", map[string]string{
		"addr": srv.Addr,
		"env":  app.config.env,
	})
	//start server as normal, return any error
	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	//otherwise wait to get return value from shutdown()
	err = <-shutdownError
	if err != nil {
		return err
	}
	//at this point we know graceful shutdown completed successfull
	app.logger.PrintInfo("stopped server", map[string]string{
		"addr": srv.Addr,
	})
	return nil

	//we are saying upon sigint or sigterm, stop accepting http req and complete all in 5s or shut it down

}
