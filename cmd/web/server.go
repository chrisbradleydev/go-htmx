package main

import (
	"context"
	"errors"
	"io/fs"
	"net"
	"net/http"
	"time"

	"github.com/chrisbradleydev/go-htmx/ui"
)

func (app *application) serve(ctx context.Context) error {
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
	mux.HandleFunc("GET /roll-d20", app.rollD20Handler)

	mux.HandleFunc("POST /contacts", app.addContact)
	mux.HandleFunc("DELETE /contacts/{id}", app.deleteContact)

	// configure server
	srv := &http.Server{
		Addr:         ":" + app.config.port,
		Handler:      mux,
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		BaseContext:  func(_ net.Listener) context.Context { return ctx },
	}

	// create shutdown channel
	shutdownErr := make(chan error)

	// start background goroutine to handle shutdown
	go func() {
		<-ctx.Done()

		app.logger.Info("shutting down server", "addr", srv.Addr, "env", app.config.env)

		ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()

		err := srv.Shutdown(ctx)
		if err != nil {
			shutdownErr <- err
		}

		app.logger.Info("completing background tasks", "addr", srv.Addr, "env", app.config.env)

		// wait for background tasks
		app.wg.Wait()
		shutdownErr <- nil
	}()

	app.logger.Info("starting server", "addr", srv.Addr, "env", app.config.env)

	// start listening for requests
	err = srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	// wait for shutdown
	err = <-shutdownErr
	if err != nil {
		return err
	}

	app.logger.Info("stopped server", "addr", srv.Addr, "env", app.config.env)

	return nil
}
