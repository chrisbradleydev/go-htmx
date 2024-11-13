package main

import (
	"html/template"
	"io/fs"
	"path/filepath"
	"time"

	"github.com/chrisbradleydev/go-htmx/ui"
)

type FormData struct {
	Errors map[string]string
	Values map[string]string
}

type Store struct {
	CurrentYear int
	FormData    FormData
	PageData    PageData
}

var Characters = []Contact{
	{
		ID:    1,
		Name:  "Luke Skywalker",
		Email: "luke_skywalker@starwars.com",
	},
	{
		ID:    2,
		Name:  "Jean-Luc Picard",
		Email: "jeanluc_picard@startrek.com",
	},
	{
		ID:    3,
		Name:  "Paul Atreides",
		Email: "paul_atreides@dune.com",
	},
}

var Items = []string{
	"Lightsaber",
	"Phaser",
	"Crysknife",
}

func newStore() Store {
	return Store{
		CurrentYear: time.Now().Year(),
		FormData: FormData{
			Errors: map[string]string{},
			Values: map[string]string{},
		},
		PageData: PageData{
			Contacts: Characters,
			Items:    Items,
		},
	}
}

func humanDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format("02 Jan 2006 at 15:04")
}

var functions = template.FuncMap{
	"humanDate": humanDate,
}

func newTemplateCache() (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}

	pages, err := fs.Glob(ui.Files, "html/pages/*.gohtml")
	if err != nil {
		return nil, err
	}

	for _, page := range pages {
		name := filepath.Base(page)

		patterns := []string{
			"html/base.gohtml",
			"html/partials/*.gohtml",
			page,
		}

		ts, err := template.New(name).
			Funcs(functions).
			ParseFS(ui.Files, patterns...)
		if err != nil {
			return nil, err
		}

		cache[name] = ts
	}

	return cache, nil
}
