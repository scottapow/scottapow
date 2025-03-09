package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/joho/godotenv"
	router "github.com/scottapow/scottapow/api"
	"github.com/scottapow/scottapow/api/auth"
	"github.com/scottapow/scottapow/api/calories"
	db "github.com/scottapow/scottapow/data"
	"github.com/scottapow/scottapow/web"
)

func main() {
	godotenv.Load()

	store, err := db.Connect()
	if err != nil {
		log.Fatal("Failed to start DB", err.Error())
	}

	err = store.Setup()
	if err != nil {
		log.Fatal("Failed to setup data store", err.Error())
	}

	web := web.NewWeb()

	a, err := auth.NewAuthProvider(store.DB)
	if err != nil {
		log.Fatal(err)
	}

	s := router.New()
	staticFilesHandler := http.StripPrefix("/web/public/", http.FileServer(http.Dir("./web/public/")))
	s.Router.Handle("/web/public/", staticFilesHandler)

	// API Handlers
	caloriesService := calories.NewCaloriesService(a)
	s.Router.HandleFunc("GET /api/calories", caloriesService.WithPermission(
		"calories_read",
		func(w http.ResponseWriter, r *http.Request) {
			claims, err := a.GetUserClaims(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusForbidden)
				return
			}
			entries, err := caloriesService.GetAllCaloriesData(r.Context(), claims.ID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			err = json.NewEncoder(w).Encode(entries)

			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		},
	))
	s.Router.HandleFunc("POST /api/calories", caloriesService.WithPermission(
		"calories_write",
		func(w http.ResponseWriter, r *http.Request) {
			err := r.ParseForm()
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			for v, i := range r.PostForm {
				log.Println(v, i)
			}
		},
	))

	// HTML Handlers
	s.Router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// error ignored because auth is not required
		claims, _ := a.GetUserClaims(r)
		fmt.Println(claims)
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
	s.Router.HandleFunc("/calories", caloriesService.WithPermission(
		"calories_read",
		func(w http.ResponseWriter, r *http.Request) {
			claims, err := a.GetUserClaims(r)
			if err != nil {
				// w.Header().Set("WWW-Authenticate", "Basic realm=\"Dev"")
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}
			entries, err := caloriesService.GetAllCaloriesData(r.Context(), claims.ID)
			fmt.Println(entries)
			web.Calories(w, r, claims)
		},
	))
	s.Router.HandleFunc("/auth/{provider}/callback", func(w http.ResponseWriter, r *http.Request) {
		// get cookie
		c, err := r.Cookie("authstate")
		if err == nil {
			fmt.Println("Cookie found:", c)
		} else {
			fmt.Println("Error retrieving cookie:", err)
		}
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
