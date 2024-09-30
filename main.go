package main

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
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
	cookieName   = "authstate"
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

		// get token and check validity
		oat, err := auth.GetToken(r.FormValue("code"))
		if err != nil || !oat.Valid() {
			fmt.Println(err)
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}

		session, err := auth.Store.Get(r, AuthCookieName)
		if err != nil {
			fmt.Println(err)
			http.Error(w, http.ErrNoCookie.Error(), 1)
		}

		// get user or create
		u, err := auth.GetUserDataFromGoogle(oat)
		if err != nil {
			fmt.Println(err)
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}

		claims := &jwt.MapClaims{
			"Email":      u.Email,
			"Firstname":  u.Firstname,
			"Surname":    u.Surname,
			"ID":         u.ID,
			"Fullname":   u.Fullname,
			"PictureURL": u.PictureURL,
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		signedJWT, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		session.Values["claims"] = signedJWT
		session.Values["gat"] = oat.AccessToken
		session.Values["expiry"] = oat.Expiry.Unix()
		session.Values["grt"] = oat.RefreshToken
		err = session.Save(r, w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Printf("jwt %+v\n", signedJWT)
		http.Redirect(w, r, "/user", http.StatusTemporaryRedirect)
	})
	// r.HandleFunc("/logout/{provider}", func(w http.ResponseWriter, r *http.Request) {
	// 	w.Header().Set("Location", "/")
	// 	w.WriteHeader(http.StatusTemporaryRedirect)
	// })
	r.HandleFunc("/auth/{provider}", func(w http.ResponseWriter, r *http.Request) {
		state, _ := RandString(32)
		fmt.Println("state", state)
		url := auth.Config.AuthCodeURL(state, oauth2.AccessTypeOnline)
		cookie := http.Cookie{
			Name:    cookieName,
			Value:   state,
			Expires: time.Now().Add(time.Hour * 1),
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
	r.HandleFunc("/user", func(w http.ResponseWriter, r *http.Request) {
		session, err := auth.Store.Get(r, AuthCookieName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			// http.Redirect(w, r, "/auth/google", http.StatusTemporaryRedirect)
		}

		expiry := session.Values["expiry"].(int64)
		claims := session.Values["claims"].(string)
		accessToken := session.Values["gat"].(string)

		if time.Now().Unix() > expiry || &accessToken == nil {
			http.Error(w, "Access Expired", http.StatusUnauthorized)
		}

		token, err := jwt.Parse(claims, func(tok *jwt.Token) (interface{}, error) {
			if _, ok := tok.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Unexpected signing method: %v", tok.Header["alg"])
			}
			return []byte(os.Getenv("JWT_SECRET")), nil
		})
		if err != nil {
			log.Fatal(err)
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			t, ok := templates["user.html"]
			if !ok {
				log.Printf("template home.html not found")
			}
			data := make(map[string]interface{})
			data["BuildId"] = staticId
			data["User"] = claims

			if err := t.Execute(w, data); err != nil {
				log.Println(err)
			}
		} else {
			fmt.Println(err)
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
