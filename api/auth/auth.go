package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const (
	oauthGoogleUrlAPI = "https://www.googleapis.com/oauth2/v2/userinfo?access_token="
	AuthCookieName    = "_oauthstate"
)

var (
	clientId     = os.Getenv("GOOGLE_KEY")
	clientSecret = os.Getenv("GOOGLE_SECRET")
)

type AuthProvider struct {
	Config *oauth2.Config
	Store  *sessions.CookieStore
}

type Claims struct {
	Email      string
	Surname    string
	Firstname  string
	ID         string
	Fullname   string
	PictureURL string
	jwt.RegisteredClaims
}

type User struct {
	Email      string `json:"email"`
	Surname    string `json:"family_name"`
	Firstname  string `json:"given_name"`
	ID         string `json:"id"`
	Fullname   string `json:"name"`
	PictureURL string `json:"picture"`
}

func NewAuthProvider() (*AuthProvider, error) {
	store := sessions.NewCookieStore([]byte(os.Getenv("SESSION_SECRET")))
	store.MaxAge(86400 * 1) // 1 day
	store.Options.Secure = true
	store.Options.SameSite = http.SameSiteStrictMode
	store.Options.Path = "/"
	store.Options.HttpOnly = true
	store.Options.Secure = true

	config := &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_KEY"),
		ClientSecret: os.Getenv("GOOGLE_SECRET"),
		RedirectURL:  os.Getenv("HOST") + "/auth/google/callback",
		Endpoint:     google.Endpoint,
		Scopes:       []string{"profile", "email"},
	}

	return &AuthProvider{
		Config: config,
		Store:  store,
	}, nil
}

func (p *AuthProvider) GetToken(code string) (*oauth2.Token, error) {
	token, err := p.Config.Exchange(context.Background(), code)
	if err != nil {
		return nil, fmt.Errorf("code exchange wrong: %s", err.Error())
	}
	return token, nil
}
func (p *AuthProvider) GetUserDataFromGoogle(token *oauth2.Token) (*User, error) {
	client := p.Config.Client(context.Background(), token)
	r, err := client.Get(oauthGoogleUrlAPI)
	defer r.Body.Close()

	if err != nil || r.StatusCode != 200 {
		return nil, fmt.Errorf("failed getting user info: %s", err.Error())
	}

	return makeUser(r)
}

func makeUser(r *http.Response) (*User, error) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	fmt.Println(string(data))
	var u User
	err = json.Unmarshal(data, &u)

	if err != nil {
		return nil, err
	}

	return &u, nil
}
