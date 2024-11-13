package main

import (
	"flag"
	"html/template"
	"log/slog"
	"os"
	"sync"
)

type config struct {
	env  string
	port string
}

type application struct {
	config        config
	logger        *slog.Logger
	store         Store
	templateCache map[string]*template.Template
	wg            sync.WaitGroup
}

func main() {
	var cfg config
	flag.StringVar(&cfg.env, "env", os.Getenv("APP_ENV"), "environment (development|production)")
	flag.StringVar(&cfg.port, "port", os.Getenv("APP_PORT"), "application port")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	templateCache, err := newTemplateCache()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	store := newStore()
	app := &application{
		config:        cfg,
		logger:        logger,
		store:         store,
		templateCache: templateCache,
	}

	err = app.serve()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}
