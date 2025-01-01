package main

import (
	"log"
	"net/http"

	"github.com/joho/godotenv"
	router "github.com/scottapow/scottapow/api"
	"github.com/scottapow/scottapow/api/auth"
	db "github.com/scottapow/scottapow/data"
	"github.com/scottapow/scottapow/web"
)

const (
	cookieName = "authstate"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	store, err := db.Connect()
	if err != nil {
		log.Fatal(err)
	}

	err = store.Setup()
	if err != nil {
		log.Fatal("Failed to setup db")
	}

	web, err := web.NewWeb()
	if err != nil {
		log.Fatal(err)
	}

	a, err := auth.NewAuthProvider(web, store.DB)
	if err != nil {
		log.Fatal(err)
	}

	s := router.New()
	staticFilesHandler := http.StripPrefix("/web/static/", http.FileServer(http.Dir("./web/static/")))
	s.Router.Handle("/web/static/", staticFilesHandler)

	// HTML Handlers

	s.Router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		web.Home(w)
	})
	s.Router.HandleFunc("/user", func(w http.ResponseWriter, r *http.Request) {
		claims, err := a.GetUserClaims(r)
		if err != nil {
			// w.Header().Set("WWW-Authenticate", "Basic realm=\"Dev"")
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		web.WriteUserTemplate(w, claims)
	})

	s.Router.HandleFunc("/auth/{provider}/callback", a.HandleLoginCallback)
	s.Router.HandleFunc("/logout/{provider}", a.HandleLogout)
	s.Router.HandleFunc("/auth/{provider}", a.HandleLogin)

	s.Run(":3000")
}
