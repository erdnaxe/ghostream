package web

import (
	"html/template"
	"log"
	"net/http"
	"os"

	"gitlab.crans.org/nounous/ghostream/internal/config"
	"gitlab.crans.org/nounous/ghostream/monitoring"
)

// Preload templates
var templates = template.Must(template.ParseGlob("web/template/*.tmpl"))

// Handle site index and viewer pages
func viewerHandler(w http.ResponseWriter, r *http.Request, cfg *config.Config) {
	// Data for template
	data := struct {
		Path string
		Cfg  *config.Config
	}{Path: r.URL.Path[1:], Cfg: cfg}

	// FIXME validation on path: https://golang.org/doc/articles/wiki/#tmp_11

	// Render template
	err := templates.ExecuteTemplate(w, "base", data)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}

	// Increment monitoring
	monitoring.ViewerServed.Inc()
}

// Auth incoming stream
func streamAuthHandler(w http.ResponseWriter, r *http.Request, cfg *config.Config) {
	// FIXME POST request only with "name" and "pass"
	// if name or pass missing => 400 Malformed request
	// else login in against LDAP or static users
	http.Error(w, "Not implemented", 400)
}

// Handle static files
// We do not use http.FileServer as we do not want directory listing
func staticHandler(w http.ResponseWriter, r *http.Request, cfg *config.Config) {
	path := "./web/" + r.URL.Path
	if f, err := os.Stat(path); err == nil && !f.IsDir() {
		http.ServeFile(w, r, path)
	} else {
		http.NotFound(w, r)
	}
}

// Closure to pass configuration
func makeHandler(fn func(http.ResponseWriter, *http.Request, *config.Config), cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fn(w, r, cfg)
	}
}

// ServeHTTP server
func ServeHTTP(cfg *config.Config) {
	// Set up HTTP router and server
	mux := http.NewServeMux()
	mux.HandleFunc("/", makeHandler(viewerHandler, cfg))
	mux.HandleFunc("/rtmp/auth", makeHandler(streamAuthHandler, cfg))
	mux.HandleFunc("/static/", makeHandler(staticHandler, cfg))
	log.Printf("Listening on http://%s/", cfg.Site.ListenAddress)
	log.Fatal(http.ListenAndServe(cfg.Site.ListenAddress, mux))
}
