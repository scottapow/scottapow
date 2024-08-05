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
	"github.com/scottapow/scottapow/services"

	_ "github.com/lib/pq"
)

const (
	layoutsDir   = "templates/layouts"
	templatesDir = "templates"
	extension    = "/*.html"
)

//go:embed templates/* templates/layouts/*
var files embed.FS

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
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
	staticId := uuid.New()
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

	// API Handlers
	r.HandleFunc("/signup", services.HandleSignup).Methods(http.MethodPost)

	fmt.Println("Listening on :3000")
	http.ListenAndServe(":3000", r)
}
