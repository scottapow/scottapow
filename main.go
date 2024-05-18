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
	// API Handlers
	r := mux.NewRouter()
	r.HandleFunc("/signup", services.HandleSignup).Methods(http.MethodPost)

	// HTML Handlers
	r.Handle("/", templ.Handler(pages.Home()))

	fmt.Println("Listening on :3000")
	http.ListenAndServe(":3000", r)
}
