package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func (app *application) indexPage(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, http.StatusOK, "home.gohtml", app.store)
}

func (app *application) contactsPage(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, http.StatusOK, "contacts.gohtml", app.store)
}

func (app *application) healthzHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, "{\"status\":\"ok\"}")
}

func (app *application) searchHandler(w http.ResponseWriter, r *http.Request) {
	time.Sleep(500 * time.Millisecond) // simulate delay
	query := strings.ToLower(r.FormValue("search"))
	chars := make([]string, len(app.store.PageData.Contacts))
	for i, contact := range app.store.PageData.Contacts {
		chars[i] = contact.Name
	}

	var matches []string
	if query != "" {
		for _, char := range chars {
			if strings.Contains(strings.ToLower(char), query) {
				matches = append(matches, char)
			}
		}
	} else {
		matches = chars
	}

	w.Header().Set("Content-Type", "text/html")
	if len(matches) == 0 {
		fmt.Fprint(w, "<ul><li>No characters found</li></ul>")
		return
	}

	fmt.Fprint(w, "<ul>")
	for _, char := range matches {
		fmt.Fprintf(w, "<li>%s</li>", char)
	}
	fmt.Fprint(w, "</ul>")
}

func (app *application) loadMoreHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, `
		<div class="item">🗡️</div>
	`)
}

type Contact struct {
	ID    int
	Name  string
	Email string
}

type Contacts []Contacts

var id int

func NewContact(id int, name, email string) Contact {
	return Contact{
		ID:    id,
		Name:  name,
		Email: email,
	}
}

func contactExists(contacts []Contact, email string) bool {
	for _, c := range contacts {
		if c.Email == email {
			return true
		}
	}
	return false
}

type PageData struct {
	Contacts []Contact
	Items    []string
}

func (d *PageData) indexOf(id int) int {
	for i, contact := range d.Contacts {
		if contact.ID == id {
			return i
		}
	}
	return -1
}

func (app *application) addContact(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "unable to parse form", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	email := r.FormValue("email")

	if name == "" {
		app.store.FormData.Errors = map[string]string{"name": "Please enter a name"}
		app.store.FormData.Values = map[string]string{"name": name, "email": email}
		app.renderPartial(w, r, 422, "contacts.gohtml", "contact-form", app.store.FormData)
		return
	}

	if email == "" {
		app.store.FormData.Errors = map[string]string{"email": "Please enter an email"}
		app.store.FormData.Values = map[string]string{"name": name, "email": email}
		app.renderPartial(w, r, 422, "contacts.gohtml", "contact-form", app.store.FormData)
		return
	}

	if contactExists(app.store.PageData.Contacts, email) {
		app.store.FormData.Errors = map[string]string{"email": "Email already exists"}
		app.store.FormData.Values = map[string]string{"name": name, "email": email}
		app.renderPartial(w, r, 422, "contacts.gohtml", "contact-form", app.store.FormData)
		return
	}
	app.clearFormData()

	id = len(app.store.PageData.Contacts) + 1
	contact := NewContact(id, name, email)
	app.store.PageData.Contacts = append(app.store.PageData.Contacts, contact)

	app.renderPartial(w, r, http.StatusOK, "contacts.gohtml", "contact-form", app.store.FormData)
	app.renderPartial(w, r, http.StatusOK, "contacts.gohtml", "oob-contact", contact)
}

func (app *application) deleteContact(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		err := fmt.Errorf("Invalid ID")
		app.serverError(w, r, err)
		return
	}

	index := app.store.PageData.indexOf(id)
	if index == -1 {
		err := fmt.Errorf("Contact not found")
		app.serverError(w, r, err)
		return
	}

	time.Sleep(500 * time.Millisecond) // simulate delay
	app.store.PageData.Contacts = append(app.store.PageData.Contacts[:index],
		app.store.PageData.Contacts[index+1:]...)
}

func isName(s string) (bool, string) {
	for _, char := range s {
		if !((char >= 'A' && char <= 'Z') ||
			(char >= 'a' && char <= 'z')) {
			return false, string(char)
		}
	}
	return true, ""
}
