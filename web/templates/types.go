package templates

import "github.com/golang-jwt/jwt/v5"

type Claims struct {
	Email      string
	Surname    string
	Firstname  string
	ID         string
	Fullname   string
	PictureURL string
	CreatedAt  string
	jwt.RegisteredClaims
}
