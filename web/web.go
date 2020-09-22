package web

import (
	"html/template"
	"log"
	"net/http"
	"os"

	"gitlab.crans.org/nounous/ghostream/internal/monitoring"
)

// Options holds web package configuration
type Options struct {
	ListenAddress string
	Name          string
	Hostname      string
	Favicon       string
	WidgetURL     string
}

// Preload templates
var templates = template.Must(template.ParseGlob("web/template/*.tmpl"))

// Handle site index and viewer pages
func viewerHandler(w http.ResponseWriter, r *http.Request, cfg *Options) {
	// Data for template
	data := struct {
		Path string
		Cfg  *Options
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
func streamAuthHandler(w http.ResponseWriter, r *http.Request, cfg *Options) {
	// FIXME POST request only with "name" and "pass"
	// if name or pass missing => 400 Malformed request
	// else login in against LDAP or static users
	http.Error(w, "Not implemented", 400)
}

// Handle static files
// We do not use http.FileServer as we do not want directory listing
func staticHandler(w http.ResponseWriter, r *http.Request, cfg *Options) {
	path := "./web/" + r.URL.Path
	if f, err := os.Stat(path); err == nil && !f.IsDir() {
		http.ServeFile(w, r, path)
	} else {
		http.NotFound(w, r)
	}
}

// Closure to pass configuration
func makeHandler(fn func(http.ResponseWriter, *http.Request, *Options), cfg *Options) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fn(w, r, cfg)
	}
}

// ServeHTTP server
func ServeHTTP(cfg *Options) {
	// Set up HTTP router and server
	mux := http.NewServeMux()
	mux.HandleFunc("/", makeHandler(viewerHandler, cfg))
	mux.HandleFunc("/rtmp/auth", makeHandler(streamAuthHandler, cfg))
	mux.HandleFunc("/static/", makeHandler(staticHandler, cfg))
	log.Printf("HTTP server listening on %s", cfg.ListenAddress)
	log.Fatal(http.ListenAndServe(cfg.ListenAddress, mux))
}