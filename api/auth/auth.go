package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/sessions"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	db "github.com/scottapow/scottapow/data"
	web "github.com/scottapow/scottapow/web"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const (
	oauthGoogleUrlAPI = "https://www.googleapis.com/oauth2/v2/userinfo?access_token="
	AuthCookieName    = "oauthstate"
	stateCookieName   = "authstate"
)

var (
	clientId     = os.Getenv("GOOGLE_KEY")
	clientSecret = os.Getenv("GOOGLE_SECRET")
)

type AuthProvider struct {
	Config   *oauth2.Config
	Store    *sessions.CookieStore
	WebStore *web.Web
	DB       *pgxpool.Pool
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

type GoogleUser struct {
	Email      string `json:"email"`
	Surname    string `json:"family_name"`
	Firstname  string `json:"given_name"`
	ID         string `json:"id"`
	Fullname   string `json:"name"`
	PictureURL string `json:"picture"`
}

var store *sessions.CookieStore

func NewAuthProvider(web *web.Web, conn *pgxpool.Pool) (*AuthProvider, error) {
	store = sessions.NewCookieStore([]byte(os.Getenv("SESSION_SECRET")))
	store.MaxAge(86400 * 1) // 1 day
	store.Options.Secure = true
	store.Options.SameSite = http.SameSiteStrictMode
	store.Options.Path = "/"
	store.Options.HttpOnly = true

	config := &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_KEY"),
		ClientSecret: os.Getenv("GOOGLE_SECRET"),
		RedirectURL:  os.Getenv("HOST") + "/auth/google/callback",
		Endpoint:     google.Endpoint,
		Scopes:       []string{"profile", "email"},
	}

	return &AuthProvider{
		Config:   config,
		Store:    store,
		WebStore: web,
		DB:       conn,
	}, nil
}

func (p *AuthProvider) GetToken(code string) (*oauth2.Token, error) {
	token, err := p.Config.Exchange(context.Background(), code)
	if err != nil {
		return nil, fmt.Errorf("code exchange wrong: %s", err.Error())
	}
	return token, nil
}
func (p *AuthProvider) GetUserDataFromGoogle(token *oauth2.Token) (*GoogleUser, error) {
	client := p.Config.Client(context.Background(), token)
	r, err := client.Get(oauthGoogleUrlAPI)
	defer r.Body.Close()

	if err != nil || r.StatusCode != 200 {
		return nil, fmt.Errorf("failed getting user info: %s", err.Error())
	}

	return makeUser(r)
}
func makeUser(r *http.Response) (*GoogleUser, error) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	fmt.Println(string(data))
	var u GoogleUser
	err = json.Unmarshal(data, &u)

	if err != nil {
		return nil, err
	}

	return &u, nil
}

func (p *AuthProvider) HandleLogin(w http.ResponseWriter, r *http.Request) {
	state, _ := RandString(32)
	fmt.Println("state", state)
	url := p.Config.AuthCodeURL(state, oauth2.AccessTypeOnline)
	cookie := http.Cookie{
		Name:    stateCookieName,
		Value:   state,
		Expires: time.Now().Add(time.Hour * 1),
	}
	http.SetCookie(w, &cookie)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	return
}

func (p *AuthProvider) HandleLoginCallback(w http.ResponseWriter, r *http.Request) {
	oauthState, _ := r.Cookie(stateCookieName)
	if r.FormValue("state") != oauthState.Value {
		// TODO: notify user and clear state cookie
		fmt.Println("invalid oauth google state")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// get token and check validity
	oat, err := p.GetToken(r.FormValue("code"))
	if err != nil || !oat.Valid() {
		fmt.Println(err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	fmt.Printf("\nexpiry %+v\n", oat.Expiry.Unix())

	session, err := p.Store.Get(r, AuthCookieName)
	if err != nil {
		fmt.Println(err)
		http.Error(w, http.ErrNoCookie.Error(), 1)
	}

	// get user or create
	googleUser, err := p.GetUserDataFromGoogle(oat)
	if err != nil {
		fmt.Println(err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// get or create user
	user, err := readOrCreateUser(r.Context(), p.DB, googleUser)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	claims := &jwt.MapClaims{
		"Email":      user.Email.String,
		"Firstname":  user.Firstname.String,
		"Surname":    user.Surname.String,
		"ID":         user.Id.String(),
		"OID":        user.Oauth_provider_id.String,
		"CreatedAt":  user.Created_at.Time.Format(time.DateTime),
		"PictureURL": user.AvatarURL.String,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedJWT, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
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

	p.WebStore.WriteUserTemplate(w, &claims)
	return
	// This doesn't work, I suppose because the request bounced to another origin
	// http.Redirect(w, r, "/user", http.StatusFound)
}

func (p *AuthProvider) HandleLogout(w http.ResponseWriter, r *http.Request) {
	// TODO: convert this to an API action so a user can logout anywhere
	w.Header().Set("Location", "/")
	http.SetCookie(w, &http.Cookie{
		Name:     AuthCookieName,
		MaxAge:   -1,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		HttpOnly: true,
	})
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	return
}

func (p *AuthProvider) GetUserClaims(r *http.Request) (jwt.MapClaims, error) {
	session, err := p.Store.Get(r, AuthCookieName)
	if err != nil {
		return nil, err
	}

	if session.IsNew {
		return nil, errors.New("User Session no available")
	}
	fmt.Printf("\n\nValues %+v\n\n", session)

	expiry, ok := session.Values["expiry"]
	if !ok || expiry == nil {
		return nil, errors.New("Failed to parse session")
	}
	expiryEpoc := expiry.(int64)
	claims := session.Values["claims"].(string)
	accessToken := session.Values["gat"].(string)

	if time.Now().Unix() > expiryEpoc || &accessToken == nil {
		return nil, errors.New("Access Expired")
	}

	token, err := jwt.Parse(claims, func(tok *jwt.Token) (interface{}, error) {
		if _, ok := tok.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", tok.Header["alg"])
		}
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		return claims, nil
	} else {
		return nil, errors.New("Invalid claims format")
	}
}

// query for the user by match with oauth id and email
// if it doesn't exist create an entry and return
func readOrCreateUser(ctx context.Context, conn *pgxpool.Pool, gu *GoogleUser) (*db.UserModel, error) {
	tx, err := conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}

	var u = &db.UserModel{}

	// Query for the user
	err = tx.QueryRow(ctx, `
		SELECT id, email, created_at, updated_at, login_at, firstname, surname, avatar_url, oauth_provider_id FROM users WHERE oauth_provider_id=$1 AND email=$2
	`, gu.ID, gu.Email).Scan(&u.Id, &u.Email, &u.Created_at, &u.Updated_at, &u.Login_at, &u.Firstname, &u.Surname, &u.AvatarURL, &u.Oauth_provider_id)

	if err != nil {
		// not returning here because it could be that the user doesn't exist yet
		// TOOO: handle a case where the user exists but there was another error in the query
		fmt.Println(err)
	}

	if u.Id.Valid && u.Login_at.Valid {
		err = tx.Commit(ctx)
		if err != nil {
			return nil, err
		}
		fmt.Println(u)
		return u, nil
	}

	// new user
	pass, err := GenerateSecurePassword()
	if err != nil {
		return nil, err
	}
	hashedPass, err := bcrypt.GenerateFromPassword(pass, bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	err = tx.QueryRow(ctx, `
		INSERT INTO users (
			id,
			password,
			email,
			firstname,
			surname,
			avatar_url,
			oauth_provider,
			oauth_provider_id
		)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7)
		RETURNING id, email, created_at, updated_at, login_at, firstname, surname, avatar_url, oauth_provider_id;
	`, hashedPass, gu.Email, gu.Firstname, gu.Surname, gu.PictureURL, "google", gu.ID,
	).Scan(&u.Id, &u.Email, &u.Created_at, &u.Updated_at, &u.Login_at, &u.Firstname, &u.Surname, &u.AvatarURL, &u.Oauth_provider_id)

	if err != nil {
		return nil, err
	}

	if !u.Id.Valid {
		tx.Rollback(ctx)
		return nil, fmt.Errorf("Unknown error creating user")
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, err
	}

	fmt.Println(u)
	return u, nil
}
