package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"slices"

	"github.com/joho/godotenv"
	router "github.com/scottapow/scottapow/api"
	"github.com/scottapow/scottapow/api/auth"
	"github.com/scottapow/scottapow/api/dumps"
	db "github.com/scottapow/scottapow/data"
	"github.com/scottapow/scottapow/web"
)

func main() {
	godotenv.Load()

	store, err := db.Connect()
	if err != nil {
		log.Fatal("Failed to start DB", err.Error())
	}

	web := web.NewWeb()

	a, err := auth.NewAuthProvider(store.DB)
	if err != nil {
		log.Fatal(err)
	}

	s := router.New()
	staticFilesHandler := http.StripPrefix("/web/public/", http.FileServer(http.Dir("./web/public/")))
	s.Router.Handle("/web/public/", staticFilesHandler)

	// API Handlers https://pkg.go.dev/net/http#hdr-Patterns-ServeMux
	dumpsService := dumps.NewDumpsService(a)
	s.Router.HandleFunc("GET /api/dumps", dumpsService.WithPermission(
		"dumps_read",
		func(w http.ResponseWriter, r *http.Request) {
			claims, err := a.GetUserClaims(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusForbidden)
				return
			}

			// make sure the user has permission to view the entries

			dumpId := r.URL.Query().Get("dumpId")
			if dumpId == "" {
				http.Error(w, "Missing query parameter dumpId", http.StatusInternalServerError)
			}

			entries, err := dumpsService.GetDumpsData(r.Context(), dumpId, claims.ID)
			if err != nil {
				fmt.Println("error here")
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			type DumpResponse struct {
				entries []db.DumpEntriesModal 
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			err = json.NewEncoder(w).Encode(DumpResponse{entries: entries})

			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		},
	))
	s.Router.HandleFunc("POST /api/dumps", dumpsService.WithPermission(
		"dumps_write",
		func(w http.ResponseWriter, r *http.Request) {
			err := r.ParseForm()
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			desc := r.FormValue("desc")

			claims, err := a.GetUserClaims(r)
			dumpId, err := dumpsService.AddDump(r.Context(), claims.ID, desc)

			http.Redirect(w, r, r.Referer()+"?active="+dumpId, http.StatusTemporaryRedirect)
		},
	))

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
	s.Router.HandleFunc("/dumps/{dumpId}", func(w http.ResponseWriter, r *http.Request) {
		claims, err := a.GetUserClaims(r)
		if err != nil || !slices.Contains(claims.Permissions, "dumps_read") {
			web.DumpsUnauthenticated(w, r, claims)
		} else {
			dumpId := r.PathValue("dumpId")
			dump, err := dumpsService.GetDump(r.Context(), claims.ID, dumpId)

			if err != nil {
				fmt.Println(err.Error())
				http.Error(w, "Could not access dump", http.StatusForbidden)
			}

			web.Dump(w, r, claims, dump)
		}
	})
	s.Router.HandleFunc("/dumps", func(w http.ResponseWriter, r *http.Request) {
		claims, err := a.GetUserClaims(r)
		if err != nil || !slices.Contains(claims.Permissions, "dumps_read") {
			web.DumpsUnauthenticated(w, r, claims)
		} else {
			dumps, _ := dumpsService.GetAllDumps(r.Context(), claims.ID)
			// TODO: handle error
			fmt.Println("dumps", dumps)
			web.DumpsAuthenticated(w, r, claims, dumps)
		}
	})
	s.Router.HandleFunc("/auth/{provider}/callback", func(w http.ResponseWriter, r *http.Request) {
		claims, err := a.HandleLoginCallback(w, r)
		// error responses are handled in HandleLoginCallback
		if err == nil && claims.ID != "" {
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		}
	})
	s.Router.HandleFunc("/logout/{provider}", a.HandleLogout)
	s.Router.HandleFunc("/auth/{provider}", a.HandleLogin)

	s.Run(":3000")
}
