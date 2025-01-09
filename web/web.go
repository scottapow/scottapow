package web

import (
	"html/template"
	"net/http"

	"github.com/a-h/templ"
	"github.com/google/uuid"
	"github.com/scottapow/scottapow/web/templates"
)

var StaticId = uuid.New().String()

type TemplateMap = map[string]*template.Template
type Web struct {
}

func NewWeb() *Web {
	return &Web{}
}

func (render *Web) Home(w http.ResponseWriter, r *http.Request) {
	templ.Handler(templates.HomePage(StaticId)).ServeHTTP(w, r)
}

func (render *Web) WriteUserTemplate(w http.ResponseWriter, r *http.Request, u templates.Claims) {
	templ.Handler(templates.UserPage(StaticId, u)).ServeHTTP(w, r)
}
