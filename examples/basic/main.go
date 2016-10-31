package main

import (
	"net/http"

	"log"

	"github.com/go-playground/livereload"
)

// run using 'go run main.go', navigate to 'http://localhost:3007/home' and
// then change the colors in 'assets/css/style.css' and watch the changes
// reload live. see advanced example for template livereloading.

var page = `<html>
	<head>
		<link type="text/css" rel="stylesheet" href="/assets/css/style.css">
	</head>
	<body>
		<p>body</p>
		<script src="/assets/js/livereload.js?host=localhost"></script>
	</body>
</html>`

func main() {

	paths := []string{"assets"}

	mappings := livereload.ReloadMapping{
		".js":  nil,
		".css": nil,
	}

	_, err := livereload.ListenAndServe(livereload.DefaultPort, paths, mappings)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/home", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(page))
	})

	fs := http.FileServer(http.Dir("assets"))
	http.Handle("/", http.StripPrefix("/assets", fs))

	http.ListenAndServe(":3007", nil)
}
