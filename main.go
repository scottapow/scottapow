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

	web := web.NewWeb()

	a, err := auth.NewAuthProvider(store.DB)
	if err != nil {
		log.Fatal(err)
	}

	s := router.New()
	staticFilesHandler := http.StripPrefix("/web/public/", http.FileServer(http.Dir("./web/public/")))
	s.Router.Handle("/web/public/", staticFilesHandler)

	// HTML Handlers

	s.Router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// error ignored because auth is not required
		claims, _ := a.GetUserClaims(r)
		web.Home(w, r, claims)
	})
	s.Router.HandleFunc("/user", func(w http.ResponseWriter, r *http.Request) {
		claims, err := a.GetUserClaims(r)
		if err != nil {
			// w.Header().Set("WWW-Authenticate", "Basic realm=\"Dev"")
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		web.WriteUserTemplate(w, r, claims)
	})
	s.Router.HandleFunc("/calories", func(w http.ResponseWriter, r *http.Request) {
		claims, _ := a.GetUserClaims(r)
		web.Calories(w, r, claims)
	})
	s.Router.HandleFunc("/auth/{provider}/callback", func(w http.ResponseWriter, r *http.Request) {
		claims, err := a.HandleLoginCallback(w, r)
		// error responses are handled in HandleLoginCallback
		if err == nil {
			web.WriteUserTemplate(w, r, claims)
		}
	})
	s.Router.HandleFunc("/logout/{provider}", a.HandleLogout)
	s.Router.HandleFunc("/auth/{provider}", a.HandleLogin)

	s.Run(":3000")
}
