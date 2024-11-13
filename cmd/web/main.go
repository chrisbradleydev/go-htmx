package main

import (
	"context"
	"flag"
	"html/template"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
)

type config struct {
	env  string
	port string
}

type application struct {
	contacts      []Contact
	config        config
	logger        *slog.Logger
	templateCache map[string]*template.Template
	mu            sync.RWMutex
	wg            sync.WaitGroup
}

func main() {
	config := newConfig()
	logger := newLogger()

	templateCache, err := newTemplateCache()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	app := &application{
		contacts:      contacts,
		config:        config,
		logger:        logger,
		templateCache: templateCache,
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	err = app.serve(ctx)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}

func newConfig() (cfg config) {
	flag.StringVar(&cfg.env, "env", os.Getenv("APP_ENV"), "environment (development|production)")
	flag.StringVar(&cfg.port, "port", os.Getenv("APP_PORT"), "application port")
	flag.Parse()
	return cfg
}

func newLogger() *slog.Logger {
	level := slog.LevelInfo
	switch strings.ToLower(os.Getenv("LOG_LEVEL")) {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
	slog.SetDefault(logger)
	return logger
}
