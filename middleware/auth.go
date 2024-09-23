package auth

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/openidConnect"
)

type Auth interface {
	SetupAuthProvider() (*goth.Provider, error)
	ValidateAuthToken(*http.Request) error
}

type AuthProvider struct {
	OIDC *openidConnect.Provider
}

func (p *AuthProvider) ValidateSession(r *http.Request) error {
	value, err := gothic.GetFromSession(p.OIDC.Name(), r)
	if err != nil {
		return err
	}
	sess, err := p.OIDC.UnmarshalSession(value)
	if err != nil {
		return err
	}
	url, err := sess.GetAuthURL()
	if err != nil {
		fmt.Println("Error getting auth url from session")
		return err
	}
	fmt.Println(url)

	params := r.URL.Query()
	if params.Encode() == "" && r.Method == "POST" {
		r.ParseForm()
		params = r.Form
	}
	_, err = sess.Authorize(p.OIDC, params)
	if err != nil {
		return err
	}
	return nil
}

func NewAuthProvider() (*AuthProvider, error) {
	store := sessions.NewCookieStore([]byte(os.Getenv("SESSION_SECRET")))
	store.MaxAge(86400 * 1) // 1 day
	store.Options.Path = "/"
	store.Options.HttpOnly = true
	store.Options.Secure = !strings.HasPrefix(os.Getenv("HOST"), "http")

	gothic.Store = store

	oidc, err := openidConnect.New(
		os.Getenv("GOOGLE_KEY"),
		os.Getenv("GOOGLE_SECRET"),
		os.Getenv("HOST")+"/auth/openid-connect/callback",
		"https://accounts.google.com/.well-known/openid-configuration",
		"openid", "profile", "email",
	)
	if err != nil {
		return nil, err
	}
	goth.UseProviders(oidc)

	return &AuthProvider{
		OIDC: oidc,
	}, nil
}
