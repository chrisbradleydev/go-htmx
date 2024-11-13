package main

import (
	"context"
	"errors"
	"io/fs"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/chrisbradleydev/go-htmx/ui"
)

func (app *application) serve() error {
	// create new ServeMux
	mux := http.NewServeMux()

	// handle static files
	fsys, err := fs.Sub(ui.Files, "static")
	if err != nil {
		return err
	}
	fileServer := http.FileServerFS(fsys)
	mux.Handle("/static/", http.StripPrefix("/static/", fileServer))

	// define routes on ServeMux
	mux.HandleFunc("GET /{$}", app.indexPage)
	mux.HandleFunc("GET /contacts", app.contactsPage)
	mux.HandleFunc("GET /healthz", app.healthzHandler)

	mux.HandleFunc("POST /search", app.searchHandler)
	mux.HandleFunc("GET /load-more", app.loadMoreHandler)

	mux.HandleFunc("POST /contacts", app.addContact)
	mux.HandleFunc("DELETE /contacts/{id}", app.deleteContact)

	// configure server
	srv := &http.Server{
		Addr:         ":" + app.config.port,
		Handler:      mux,
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	shutdownError := make(chan error)

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		s := <-quit

		app.logger.Info("caught signal", "signal", s.String())

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		err := srv.Shutdown(ctx)
		if err != nil {
			shutdownError <- err
		}

		app.logger.Info("completing background tasks", "addr", srv.Addr)

		app.wg.Wait()
		shutdownError <- nil
	}()

	app.logger.Info("starting server", "addr", srv.Addr, "env", app.config.env)

	err = srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	err = <-shutdownError
	if err != nil {
		return err
	}

	app.logger.Info("stopped server", "addr", srv.Addr)

	return nil
}
