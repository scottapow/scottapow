package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

const oauthGoogleUrlAPI = "https://www.googleapis.com/oauth2/v2/userinfo?access_token="

var (
	clientId     = os.Getenv("GOOGLE_KEY")
	clientSecret = os.Getenv("GOOGLE_SECRET")
)

type AuthProvider struct {
	Config *oauth2.Config
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
	ctx := context.Background()
	// store := sessions.NewCookieStore([]byte(os.Getenv("SESSION_SECRET")))
	// store.MaxAge(86400 * 1) // 1 day
	// store.Options.Secure = true
	// store.Options.SameSite = http.SameSiteStrictMode
	// store.Options.Path = "/"
	// store.Options.HttpOnly = true
	// store.Options.Secure = true

	provider, err := oidc.NewProvider(ctx, "https://accounts.google.com")
	if err != nil {
		return nil, err
	}
	oauth2Config := &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_KEY"),
		ClientSecret: os.Getenv("GOOGLE_SECRET"),
		RedirectURL:  os.Getenv("HOST") + "/auth/google/callback",
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{"profile", "email"},
	}

	return &AuthProvider{
		Config: oauth2Config,
	}, nil
}

func (p *AuthProvider) GetUserDataFromGoogle(code string) (*User, error) {
	// Use code to get token and get user info from Google.

	token, err := p.Config.Exchange(context.Background(), code)
	if err != nil {
		return nil, fmt.Errorf("code exchange wrong: %s", err.Error())
	}
	client := p.Config.Client(context.Background(), token)
	r, err := client.Get(oauthGoogleUrlAPI)
	if err != nil {
		return nil, fmt.Errorf("failed getting user info: %s", err.Error())
	}

	defer r.Body.Close()

	if err != nil {
		return nil, fmt.Errorf("failed read response: %s", err.Error())
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
