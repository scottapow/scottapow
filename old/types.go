package main

type GoogleUser struct {
	Email      string `json:"email"`
	Surname    string `json:"family_name"`
	Firstname  string `json:"given_name"`
	ID         string `json:"id"`
	Fullname   string `json:"name"`
	PictureURL string `json:"picture"`
}
