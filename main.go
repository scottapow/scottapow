package main

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/markbates/goth/gothic"
	auth "github.com/scottapow/scottapow/middleware"
	"github.com/scottapow/scottapow/services"
)

const (
	layoutsDir   = "templates/layouts"
	templatesDir = "templates"
	extension    = "/*.html"
)

var staticId = uuid.New()

//go:embed templates/* templates/layouts/*
var files embed.FS

func main() {
	godotenv.Load()

	provider, err := auth.NewAuthProvider()
	if err != nil {
		log.Fatal(err)
	}

	templates := make(map[string]*template.Template)
	tmplFiles, err := fs.ReadDir(files, templatesDir)
	if err != nil {
		log.Fatal(err)
	}
	for _, tmpl := range tmplFiles {
		if tmpl.IsDir() {
			continue
		}

		pt, err := template.ParseFS(files, templatesDir+"/"+tmpl.Name(), layoutsDir+extension)
		if err != nil {
			log.Fatal(err)
		}

		templates[tmpl.Name()] = pt
	}

	r := mux.NewRouter()

	// Include all static files
	s := http.StripPrefix("/static/", http.FileServer(http.Dir("./static/")))
	r.PathPrefix("/static/").Handler(s)

	// HTML Handlers
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		t, ok := templates["home.html"]
		if !ok {
			log.Printf("template home.html not found")
		}
		data := make(map[string]interface{})
		data["BuildId"] = staticId

		if err := t.Execute(w, data); err != nil {
			log.Println(err)
		}
	})

	r.HandleFunc("/protected", func(w http.ResponseWriter, r *http.Request) {
		err := provider.ValidateSession(r)
		if err != nil {
			http.Redirect(w, r, "/auth/"+provider.OIDC.Name(), http.StatusTemporaryRedirect)
		}

		t, ok := templates["protected.html"]
		if !ok {
			log.Printf("template protected.html not found")
		}
		data := make(map[string]interface{})
		data["BuildId"] = staticId

		if err := t.Execute(w, data); err != nil {
			log.Println(err)
		}
	})

	r.HandleFunc("/auth/{provider}/callback", func(w http.ResponseWriter, r *http.Request) {
		t, ok := templates["user.html"]
		if !ok {
			log.Printf("template user.html not found")
		}
		data := make(map[string]interface{})
		data["BuildId"] = staticId
		user, err := gothic.CompleteUserAuth(w, r)
		if err != nil {
			fmt.Fprintln(w, err)
			return
		}
		data["User"] = user
		if err := t.Execute(w, data); err != nil {
			log.Println(err)
		}
	})
	r.HandleFunc("/logout/{provider}", func(w http.ResponseWriter, r *http.Request) {
		gothic.Logout(w, r)
		w.Header().Set("Location", "/")
		w.WriteHeader(http.StatusTemporaryRedirect)
	})
	r.HandleFunc("/auth/{provider}", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r.Host)
		fmt.Println(r)
		// try to get the user without re-authenticating
		if user, err := gothic.CompleteUserAuth(w, r); err == nil {
			t, ok := templates["user.html"]
			if !ok {
				log.Printf("template user.html not found")
			}
			fmt.Println(user)
			data := make(map[string]interface{})
			data["BuildId"] = staticId
			data["User"] = user
			if err := t.Execute(w, data); err != nil {
				log.Println(err)
			}
		} else {
			gothic.BeginAuthHandler(w, r)
		}
	})

	// API Handlers
	r.HandleFunc("/signup", services.HandleSignup).Methods(http.MethodPost)

	fmt.Println("Listening on :3000")
	http.ListenAndServe(":3000", r)
}
