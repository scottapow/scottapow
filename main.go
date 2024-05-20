package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/scottapow/scottapow/pages"
	"github.com/scottapow/scottapow/services"

	"github.com/a-h/templ"
)

func main() {
	r := mux.NewRouter()

	// HTML Handlers
	r.Handle("/", templ.Handler(pages.Home()))

	// API Handlers
	r.HandleFunc("/signup", services.HandleSignup).Methods(http.MethodPost)

	fmt.Println("Listening on :3000")
	http.ListenAndServe(":3000", r)
}
