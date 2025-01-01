package web

import (
	"embed"
	"html/template"
	"io/fs"
	"log"
	"net/http"

	"github.com/google/uuid"
)

const (
	layoutsDir   = "templates/layouts"
	templatesDir = "templates"
	extension    = "/*.html"
	cookieName   = "authstate"
)

//go:embed templates/* templates/layouts/*
var files embed.FS
var StaticId = uuid.New()

type TemplateMap = map[string]*template.Template
type Web struct {
	Templates TemplateMap
}

func NewWeb() (*Web, error) {
	templates := make(TemplateMap)
	tmplFiles, err := fs.ReadDir(files, templatesDir)
	if err != nil {
		return nil, err
	}

	for _, tmpl := range tmplFiles {
		if tmpl.IsDir() {
			continue
		}

		pt, err := template.ParseFS(files, templatesDir+"/"+tmpl.Name(), layoutsDir+extension)
		if err != nil {
			return nil, err
		}

		templates[tmpl.Name()] = pt
	}

	return &Web{Templates: templates}, nil
}

func (render *Web) Home(w http.ResponseWriter) {
	t, ok := render.Templates["home.html"]
	if !ok {
		log.Printf("template home.html not found")
	}
	data := make(map[string]interface{})
	data["BuildId"] = StaticId

	if err := t.Execute(w, data); err != nil {
		log.Println(err)
	}
}

func (render *Web) WriteUserTemplate(w http.ResponseWriter, u any) {
	t, ok := render.Templates["user.html"]
	if !ok {
		log.Printf("template home.html not found")
	}
	data := make(map[string]interface{})
	data["BuildId"] = StaticId
	data["User"] = u

	if err := t.Execute(w, data); err != nil {
		log.Println(err)
	}
}

// put all template shit here
