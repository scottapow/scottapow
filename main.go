package main

import (
	"database/sql"
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/openidConnect"
	"github.com/scottapow/scottapow/services"

	_ "github.com/lib/pq"
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
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	// goth.UseProviders(
	// 	google.New(os.Getenv("GOOGLE_KEY"), os.Getenv("GOOGLE_SECRET"), "http://localhost:3000/auth/google/callback"),
	// )

	openidConnect, _ := openidConnect.New(os.Getenv("GOOGLE_KEY"), os.Getenv("GOOGLE_SECRET"), "http://localhost:3000/auth/openid-connect/callback", "https://accounts.google.com/.well-known/openid-configuration")
	if openidConnect != nil {
		goth.UseProviders(openidConnect)
	}

	connStr := os.Getenv("DB_CONN_STR")
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	rows, err := db.Query("select version()")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var version string
	for rows.Next() {
		err := rows.Scan(&version)
		if err != nil {
			log.Fatal(err)
		}
	}
	fmt.Printf("version=%s\n", version)


	env := os.Getenv("APP_ENV")
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
		data["Env"] = env

		if err := t.Execute(w, data); err != nil {
			log.Println(err)
		}
	})

	if env == "development" {
		r.HandleFunc("/auth/{provider}/callback", func(w http.ResponseWriter, r *http.Request) {
			t, ok := templates["user.html"]
			if !ok {
				log.Printf("template user.html not found")
			}
			data := make(map[string]interface{})
			data["BuildId"] = staticId
			data["Env"] = env
			user, err := gothic.CompleteUserAuth(w, r)
			if err != nil {
				fmt.Fprintln(w, err)
				return
			}
			fmt.Println(user)
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
			// try to get the user without re-authenticating
			if user, err := gothic.CompleteUserAuth(w, r); err == nil {
				t, ok := templates["user.html"]
				if !ok {
					log.Printf("template user.html not found")
				}
				fmt.Println(user)
				data := make(map[string]interface{})
				data["BuildId"] = staticId
				data["Env"] = env
				data["User"] = user
				if err := t.Execute(w, data); err != nil {
					log.Println(err)
				}
			} else {
				gothic.BeginAuthHandler(w, r)
			}
		})
	}

	// API Handlers
	r.HandleFunc("/signup", services.HandleSignup).Methods(http.MethodPost)

	fmt.Println("Listening on :3000")
	http.ListenAndServe(":3000", r)
}
