package web

import (
	"html/template"
	"log"
	"net/http"
	"os"

	"gitlab.crans.org/nounous/ghostream/internal/config"
)

// Preload templates
var templates = template.Must(template.ParseGlob("web/template/*.tmpl"))

// Handle site index and viewer pages
func handlerViewer(w http.ResponseWriter, r *http.Request) {
	// Remove traling slash
	//path := r.URL.Path[1:]

	// Render template
	err := templates.ExecuteTemplate(w, "base", nil)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// Auth incoming stream
func handleStreamAuth(w http.ResponseWriter, r *http.Request) {
	// FIXME POST request only with "name" and "pass"
	// if name or pass missing => 400 Malformed request
	// else login in against LDAP or static users
	http.Error(w, "Not implemented", 400)
}

// Handle static files
// We do not use http.FileServer as we do not want directory listing
func handleStatic(w http.ResponseWriter, r *http.Request) {
	path := "./web/" + r.URL.Path
	if f, err := os.Stat(path); err == nil && !f.IsDir() {
		http.ServeFile(w, r, path)
	} else {
		http.NotFound(w, r)
	}
}

// ServeHTTP server
func ServeHTTP(cfg *config.Config) {
	// Set up HTTP router and server
	http.HandleFunc("/", handlerViewer)
	http.HandleFunc("/rtmp/auth", handleStreamAuth)
	http.HandleFunc("/static/", handleStatic)
	log.Printf("Listening on http://%s/", cfg.Site.ListenAdress)
	log.Fatal(http.ListenAndServe(cfg.Site.ListenAdress, nil))
}
