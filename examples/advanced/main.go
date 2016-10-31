package main

import (
	"html/template"
	"net/http"

	"log"

	"github.com/go-playground/livereload"
)

// run using 'go run main.go', navigate to 'http://localhost:3007/home' and
// then change the colors in 'assets/css/style.css' or 'templates/home./tmpl'
// and watch the changes reload live.

var (
	templates *template.Template
)

func main() {

	var err error

	templates, err = compileTemplates()
	if err != nil {
		log.Fatal(err)
	}

	paths := []string{
		"assets",
		"templates",
	}

	tmplFn := func(name string) (bool, error) {
		tmpls, err := compileTemplates()
		if err != nil {
			return false, err
		}

		*templates = *tmpls

		return true, nil
	}

	mappings := livereload.ReloadMapping{
		".js":   nil,
		".css":  nil,
		".tmpl": tmplFn,
	}

	_, err = livereload.ListenAndServe(livereload.DefaultPort, paths, mappings)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/home", func(w http.ResponseWriter, r *http.Request) {
		err := templates.ExecuteTemplate(w, "home", nil)
		if err != nil {
			log.Fatal(err)
		}
	})

	fs := http.FileServer(http.Dir("assets"))
	http.Handle("/", http.StripPrefix("/assets", fs))

	http.ListenAndServe(":3007", nil)
}

func compileTemplates() (*template.Template, error) {
	return template.ParseGlob("templates/*.tmpl")
}
