package web

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/google/uuid"
	db "github.com/scottapow/scottapow/data"
	"github.com/scottapow/scottapow/web/templates"
)

var StaticId = uuid.New().String()

type Web struct {
}

func NewWeb() *Web {
	return &Web{}
}

func (render *Web) Home(w http.ResponseWriter, r *http.Request, u templates.Claims) {
	templ.Handler(templates.HomePage(StaticId, u)).ServeHTTP(w, r)
}

func (render *Web) WriteUserTemplate(w http.ResponseWriter, r *http.Request, u templates.Claims) {
	templ.Handler(templates.UserPage(StaticId, u)).ServeHTTP(w, r)
}

func (render *Web) DumpsAuthenticated(w http.ResponseWriter, r *http.Request, u templates.Claims, dumps []db.DumpsModel) {
	templ.Handler(templates.DumpsAuthenticatedPage(StaticId, u, dumps)).ServeHTTP(w, r)
}

func (render *Web) DumpsUnauthenticated(w http.ResponseWriter, r *http.Request, u templates.Claims) {
	templ.Handler(templates.DumpsUnAuthenticatedPage(StaticId, u)).ServeHTTP(w, r)
}

func (render *Web) Dump(w http.ResponseWriter, r *http.Request, u templates.Claims, d db.DumpsModel) {
	templ.Handler(templates.DumpPage(StaticId, u, d)).ServeHTTP(w, r)
}
