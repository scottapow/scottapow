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
	"github.com/scottapow/scottapow/web/templates"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const (
	oauthGoogleUrlAPI = "https://www.googleapis.com/oauth2/v2/userinfo?access_token="
	AuthCookieName    = "oauthstate"
	stateCookieName   = "__session"
)

var (
	clientId     = os.Getenv("GOOGLE_KEY")
	clientSecret = os.Getenv("GOOGLE_SECRET")
)

type AuthProvider struct {
	Config *oauth2.Config
	Store  *sessions.CookieStore
	DB     *pgxpool.Pool
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

func NewAuthProvider(conn *pgxpool.Pool) (*AuthProvider, error) {
	store = sessions.NewCookieStore([]byte(os.Getenv("SESSION_SECRET")))
	store.MaxAge(86400 * 1) // 1 day
	store.Options.Secure = true
	store.Options.SameSite = http.SameSiteNoneMode
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
		Config: config,
		Store:  store,
		DB:     conn,
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

	if err != nil || r.StatusCode != 200 {
		return nil, fmt.Errorf("failed getting user info: %s", err.Error())
	}

	defer r.Body.Close()
	return makeUser(r)
}
func makeUser(r *http.Response) (*GoogleUser, error) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	var u GoogleUser
	err = json.Unmarshal(data, &u)

	if err != nil {
		return nil, err
	}

	return &u, nil
}

func (p *AuthProvider) HandleLogin(w http.ResponseWriter, r *http.Request) {
	state, _ := RandString(32)
	url := p.Config.AuthCodeURL(state, oauth2.AccessTypeOnline)
	cookie := http.Cookie{
		Name:     stateCookieName,
		Value:    state,
		Expires:  time.Now().Add(time.Hour),
		Path:     "/",
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
		HttpOnly: true,
	}
	http.SetCookie(w, &cookie)
	r.AddCookie(&cookie)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (p *AuthProvider) HandleLoginCallback(w http.ResponseWriter, r *http.Request) (templates.Claims, error) {
	oauthState, err := r.Cookie(stateCookieName)
	if err != nil {
		http.Error(w, http.ErrNoCookie.Error(), http.StatusInternalServerError)
		return templates.Claims{}, err
	}
	if r.FormValue("state") != oauthState.Value {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return templates.Claims{}, errors.New("Invalid State")
	}

	// get token and check validity
	oat, err := p.GetToken(r.FormValue("code"))
	if err != nil || !oat.Valid() {
		fmt.Println(err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return templates.Claims{}, err
	}
	fmt.Printf("\nexpiry %+v\n", oat.Expiry.Unix())

	session, err := p.Store.Get(r, AuthCookieName)
	if err != nil {
		fmt.Println(err)
		http.Error(w, http.ErrNoCookie.Error(), http.StatusInternalServerError)
		return templates.Claims{}, err
	}

	// get user or create
	googleUser, err := p.GetUserDataFromGoogle(oat)
	if err != nil {
		fmt.Println(err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return templates.Claims{}, err
	}

	// get or create user
	user, err := readOrCreateUser(r.Context(), p.DB, googleUser)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return templates.Claims{}, err
	}

	permissions, err := readPermissions(r.Context(), p.DB, user.Id.String())

	claims := &jwt.MapClaims{
		"Email":       user.Email.String,
		"Firstname":   user.Firstname.String,
		"Surname":     user.Surname.String,
		"ID":          user.Id.String(),
		"OID":         user.Oauth_provider_id.String,
		"CreatedAt":   user.Created_at.Time.Format(time.DateTime),
		"PictureURL":  user.AvatarURL.String,
		"Permissions": permissions,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedJWT, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return templates.Claims{}, err
	}

	session.Values["claims"] = signedJWT
	session.Values["gat"] = oat.AccessToken
	session.Values["expiry"] = oat.Expiry.Unix()
	session.Values["grt"] = oat.RefreshToken
	err = session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return templates.Claims{}, err
	}

	return templates.Claims{
		Email:       user.Email.String,
		Firstname:   user.Firstname.String,
		Surname:     user.Surname.String,
		ID:          user.Id.String(),
		PictureURL:  user.AvatarURL.String,
		CreatedAt:   user.Created_at.Time.Format(time.DateTime),
		Permissions: permissions,
	}, nil
	// This doesn't work, I suppose because the request bounced to another origin
	// http.Redirect(w, r, "/user", http.StatusFound)
}

func readPermissions(ctx context.Context, conn *pgxpool.Pool, userId string) ([]string, error) {
	fmt.Println("reading permissions")
	rows, err := conn.Query(ctx, `SELECT * from permissions WHERE user_id=$1`, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []string
	for rows.Next() {
		var p db.PermissionModel
		err = rows.Scan(&p.Id, &p.User_id, &p.Permission, &p.Created_at)
		if err != nil {
			fmt.Println(err.Error())
			return nil, err
		}
		permissions = append(permissions, p.Permission.String)
	}

	return permissions, nil
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
}

func (p *AuthProvider) GetUserClaims(r *http.Request) (templates.Claims, error) {
	session, err := p.Store.Get(r, AuthCookieName)
	var templateClaims = templates.Claims{}
	if err != nil {
		return templateClaims, err
	}

	if session.IsNew {
		return templateClaims, errors.New("User Session not available")
	}

	expiry, ok := session.Values["expiry"]
	if !ok || expiry == nil {
		return templateClaims, errors.New("Failed to parse session")
	}
	expiryEpoc := expiry.(int64)
	claims := session.Values["claims"].(string)
	accessToken := session.Values["gat"].(string)

	if time.Now().Unix() > expiryEpoc || accessToken == "" {
		return templateClaims, errors.New("Access Expired")
	}

	token, err := jwt.Parse(claims, func(tok *jwt.Token) (interface{}, error) {
		if _, ok := tok.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", tok.Header["alg"])
		}
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	if err != nil {
		return templateClaims, err
	}

	c, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return templateClaims, errors.New("Invalid claims format")
	}

	templateClaims.Email, ok = c["Email"].(string)
	templateClaims.Firstname, ok = c["Firstname"].(string)
	templateClaims.Surname, ok = c["Surname"].(string)
	templateClaims.ID, ok = c["ID"].(string)
	templateClaims.PictureURL, ok = c["PictureURL"].(string)
	templateClaims.CreatedAt, ok = c["CreatedAt"].(string)
	permissions, ok := c["Permissions"].([]interface{})
	if ok && len(permissions) > 0 {
		templateClaims.Permissions = make([]string, len(permissions))
		for i, v := range permissions {
			templateClaims.Permissions[i] = v.(string)
		}
	}

	return templateClaims, nil
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
	row := tx.QueryRow(ctx, `
		SELECT id, email, created_at, updated_at, login_at, firstname, surname, avatar_url, oauth_provider_id FROM users WHERE oauth_provider_id=$1 AND email=$2
	`, gu.ID, gu.Email)
	row.Scan()
	.Scan(&pgu.Id, &pgu.Email, &pgu.Created_at, &pgu.Updated_at, &pgu.Login_at, &pgu.Firstname, &pgu.Surname, &pgu.AvatarURL, &pgu.Oauth_provider_id)

	if err != nil {
		tx.Rollback(ctx)
		// no user exist. Not creating new users now
		// TODO: add a check against email whitelist
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("User is not allowed")
		} else {
			return nil, fmt.Errorf("Error reading user: %w", err)
		}
	}

	if pgu.Id.Valid && pgu.Login_at.Valid {
		err = tx.Commit(ctx)
		if err != nil {
			return nil, err
		}
		
		u := createUserFromDB(*pgu)
		return &u, nil
	}

	// new user
	pass, err := GenerateSecurePassword()
	if err != nil {
		tx.Rollback(ctx)
		return nil, err
	}
	hashedPass, err := bcrypt.GenerateFromPassword(pass, bcrypt.DefaultCost)
	if err != nil {
		tx.Rollback(ctx)
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
	).Scan(&pgu.Id, &pgu.Email, &pgu.Created_at, &pgu.Updated_at, &pgu.Login_at, &pgu.Firstname, &pgu.Surname, &pgu.AvatarURL, &pgu.Oauth_provider_id)

	if err != nil {
		tx.Rollback(ctx)
		return nil, err
	}

	r, err := tx.Query(ctx, `
		INSERT INTO permissions (id, user_id, permission)
		VALUES (gen_random_uuid(), $1, 'dumps_read'), (gen_random_uuid(), $1, 'dumps_write');
	`, Id.String())
	r.Close()

	if err != nil {
		tx.Rollback(ctx)
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

	return u, nil
}

func createUserFromDB(pgu db.PGUserModel) db.UserModel {
	u := db.UserModel{
		Id: pgu.Id.String(),
		Email: pgu.Email.String,
		Created_at: pgu.Created_at.Time,
		Updated_at: pgu.Updated_at.Time,
		Login_at: pgu.Login_at.Time,
		Firstname: pgu.Firstname.String,
		Surname: pgu.Surname.String,
		AvatarURL: pgu.AvatarURL.String,
		Oauth_provider_id: pgu.Oauth_provider_id.String,
	}
	return u
}