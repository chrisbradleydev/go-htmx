package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type FormData struct {
	Errors map[string]string
	Values map[string]string
}

type PageData struct {
	Contacts []Contact
}

type RequestData struct {
	CurrentYear int
	FormData    FormData
	PageData    PageData
}

func newRequestData() RequestData {
	return RequestData{
		CurrentYear: time.Now().Year(),
		FormData: FormData{
			Errors: map[string]string{},
			Values: map[string]string{},
		},
		PageData: PageData{
			Contacts: contacts,
		},
	}
}

var contacts = []Contact{
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

func (app *application) indexPage(w http.ResponseWriter, r *http.Request) {
	data := newRequestData()
	data.PageData.Contacts = app.getContacts()
	app.render(w, r, http.StatusOK, "home.gohtml", &data)
}

func (app *application) contactsPage(w http.ResponseWriter, r *http.Request) {
	data := newRequestData()
	data.PageData.Contacts = app.getContacts()
	app.render(w, r, http.StatusOK, "contacts.gohtml", &data)
}

func (app *application) healthzHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, "{\"status\":\"ok\"}")
}

func (app *application) searchHandler(w http.ResponseWriter, r *http.Request) {
	time.Sleep(500 * time.Millisecond) // simulate delay
	query := strings.ToLower(r.FormValue("search"))

	names := app.contactNames()

	var matchedNames []string
	if query != "" {
		for _, n := range names {
			if strings.Contains(strings.ToLower(n), query) {
				matchedNames = append(matchedNames, n)
			}
		}
	} else {
		matchedNames = names
	}

	w.Header().Set("Content-Type", "text/html")
	if len(matchedNames) == 0 {
		fmt.Fprint(w, "<ul><li>No contacts found</li></ul>")
		return
	}

	fmt.Fprint(w, "<ul>")
	for _, n := range matchedNames {
		fmt.Fprintf(w, "<li>%s</li>", n)
	}
	fmt.Fprint(w, "</ul>")
}

type Contact struct {
	ID    int
	Name  string
	Email string
}

type Contacts []Contact

func newContact(id int, name, email string) Contact {
	return Contact{
		ID:    id,
		Name:  name,
		Email: email,
	}
}

func (app *application) contactExists(email string) bool {
	app.mu.RLock()
	defer app.mu.RUnlock()

	for _, c := range app.contacts {
		if c.Email == email {
			return true
		}
	}
	return false
}

func (app *application) contactIndex(id int) int {
	app.mu.RLock()
	defer app.mu.RUnlock()

	for i, c := range app.contacts {
		if c.ID == id {
			return i
		}
	}
	return -1
}

func (app *application) contactNames() []string {
	app.mu.RLock()
	defer app.mu.RUnlock()

	names := make([]string, len(app.contacts))
	for i, c := range app.contacts {
		names[i] = c.Name
	}
	return names
}

func (app *application) getContacts() []Contact {
	app.mu.RLock()
	defer app.mu.RUnlock()
	return app.contacts
}

func (app *application) addContact(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "unable to parse form", http.StatusBadRequest)
		return
	}

	data := newRequestData()
	name := r.FormValue("name")
	email := r.FormValue("email")

	if name == "" {
		data.FormData.Errors = map[string]string{"name": "Please enter a name"}
		data.FormData.Values = map[string]string{"name": name, "email": email}
		app.renderPartial(w, r, 422, "contacts.gohtml", "contact-form", data.FormData)
		return
	}

	if email == "" {
		data.FormData.Errors = map[string]string{"email": "Please enter an email"}
		data.FormData.Values = map[string]string{"name": name, "email": email}
		app.renderPartial(w, r, 422, "contacts.gohtml", "contact-form", data.FormData)
		return
	}

	if app.contactExists(email) {
		data.FormData.Errors = map[string]string{"email": "Email already exists"}
		data.FormData.Values = map[string]string{"name": name, "email": email}
		app.renderPartial(w, r, 422, "contacts.gohtml", "contact-form", data.FormData)
		return
	}

	app.mu.Lock()
	contact := newContact(len(app.contacts)+1, name, email)
	app.contacts = append(app.contacts, contact)
	app.mu.Unlock()

	app.renderPartial(w, r, http.StatusOK, "contacts.gohtml", "contact-form", data.FormData)
	app.renderPartial(w, r, http.StatusOK, "contacts.gohtml", "oob-contact", contact)
}

func (app *application) deleteContact(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		err := fmt.Errorf("invalid ID")
		app.serverError(w, r, err)
		return
	}

	index := app.contactIndex(id)
	if index == -1 {
		err := fmt.Errorf("Contact not found")
		app.serverError(w, r, err)
		return
	}

	time.Sleep(500 * time.Millisecond) // simulate delay

	app.mu.Lock()
	defer app.mu.Unlock()
	app.contacts = append(app.contacts[:index],
		app.contacts[index+1:]...)
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

func (app *application) rollD20Handler(w http.ResponseWriter, r *http.Request) {
	roll := 1 + rand.Intn(20)
	rollStr := strconv.Itoa(roll)

	w.Header().Set("Content-Type", "text/html")
	if roll != 20 {
		fmt.Fprintf(w, `<div>%s</div>`, rollStr)
	} else {
		fmt.Fprintf(w, `<div class="error">%s</div>`, rollStr)
	}
}
