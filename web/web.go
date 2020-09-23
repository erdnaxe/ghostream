package web

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/pion/webrtc/v3"
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
var templates = template.Must(template.ParseGlob("web/template/*.html"))

// Handle WebRTC session description exchange via POST
func sessionExchangeHandler(w http.ResponseWriter, r *http.Request) {
	// Limit response body to 128KB
	r.Body = http.MaxBytesReader(w, r.Body, 131072)

	// Decode client description
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	remoteDescription := webrtc.SessionDescription{}
	if err := dec.Decode(&remoteDescription); err != nil {
		http.Error(w, "The JSON WebRTC offer is malformed", http.StatusBadRequest)
		return
	}

	// FIXME remoteDescription -> "Magic" -> localDescription
	localDescription := remoteDescription

	// Send server description as JSON
	jsonDesc, err := json.Marshal(localDescription)
	if err != nil {
		http.Error(w, "An error occured while formating response", http.StatusInternalServerError)
		log.Println("An error occured while sending session description", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonDesc)
}

// Handle site index and viewer pages
// POST requests are used to exchange WebRTC session descriptions
func viewerHandler(w http.ResponseWriter, r *http.Request, cfg *Options) {
	// FIXME validation on path: https://golang.org/doc/articles/wiki/#tmp_11

	switch r.Method {
	case "GET":
		// Render template
		data := struct {
			Path string
			Cfg  *Options
		}{Path: r.URL.Path[1:], Cfg: cfg}
		if err := templates.ExecuteTemplate(w, "base", data); err != nil {
			log.Println(err.Error())
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	case "POST":
		sessionExchangeHandler(w, r)
	default:
		http.Error(w, "Sorry, only GET and POST methods are supported.", http.StatusBadRequest)
	}

	// Increment monitoring
	monitoring.ViewerServed.Inc()
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
	mux.HandleFunc("/static/", makeHandler(staticHandler, cfg))
	log.Printf("HTTP server listening on %s", cfg.ListenAddress)
	log.Fatal(http.ListenAndServe(cfg.ListenAddress, mux))
}
