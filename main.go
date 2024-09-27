package main

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/scottapow/scottapow/services"
	"golang.org/x/oauth2"
)

const (
	layoutsDir   = "templates/layouts"
	templatesDir = "templates"
	extension    = "/*.html"
	cookieName   = "_oauthstate"
)

var staticId = uuid.New()

//go:embed templates/* templates/layouts/*
var files embed.FS

func main() {
	godotenv.Load()

	auth, err := NewAuthProvider()
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
	r.HandleFunc("/auth/{provider}/callback", func(w http.ResponseWriter, r *http.Request) {
		oauthState, _ := r.Cookie(cookieName)
		if r.FormValue("state") != oauthState.Value {
			fmt.Println("invalid oauth google state")
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}

		user, err := auth.GetUserDataFromGoogle(r.FormValue("code"))
		if err != nil {
			fmt.Println(err)
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}

		// GetOrCreate User in your db.
		// Redirect or response with a token.
		// More code .....
		fmt.Printf("%+v\n", user)

		if err != nil {
			fmt.Fprintln(w, err)
			return
		}
		t, ok := templates["user.html"]
		if !ok {
			log.Printf("template user.html not found")
		}
		data := make(map[string]interface{})
		data["BuildId"] = staticId
		data["User"] = user
		t.Execute(w, data)
	})
	// r.HandleFunc("/logout/{provider}", func(w http.ResponseWriter, r *http.Request) {
	// 	w.Header().Set("Location", "/")
	// 	w.WriteHeader(http.StatusTemporaryRedirect)
	// })
	r.HandleFunc("/auth/{provider}", func(w http.ResponseWriter, r *http.Request) {
		state, _ := RandString(32)
		url := auth.Config.AuthCodeURL(state, oauth2.AccessTypeOnline)
		cookie := http.Cookie{
			Name:    cookieName,
			Value:   state,
			Expires: time.Now().Add(time.Hour * 24 * 30),
		}
		http.SetCookie(w, &cookie)
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	})
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
	// r.HandleFunc("/protected", func(w http.ResponseWriter, r *http.Request) {
	// 	err := provider.ValidateSession(r)
	// 	if err != nil {
	// 		http.Redirect(w, r, "/auth/"+provider.OIDC.Name(), http.StatusTemporaryRedirect)
	// 	}

	// 	t, ok := templates["protected.html"]
	// 	if !ok {
	// 		log.Printf("template protected.html not found")
	// 	}
	// 	data := make(map[string]interface{})
	// 	data["BuildId"] = staticId
	// 	t.Execute(w, data)
	// })

	// API Handlers
	r.HandleFunc("/signup", services.HandleSignup).Methods(http.MethodPost)

	fmt.Println("Listening on :3000")
	http.ListenAndServe(":3000", r)
}
