package main

import (
	"fmt"
	"net/http"

	"github.com/scottapow/scottapow/pages"

	"github.com/a-h/templ"
)

func main() {
	http.Handle("/", templ.Handler(pages.Home()))

	fmt.Println("Listening on :3000")
	http.ListenAndServe(":3000", nil)
}
