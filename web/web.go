package web

import (
	"encoding/json"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/markbates/pkger"
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

var (
	cfg *Options

	// WebRTC session description channels
	remoteSdpChan chan webrtc.SessionDescription
	localSdpChan  chan webrtc.SessionDescription

	// Preload templates
	templates *template.Template

	// Precompile regex
	validPath = regexp.MustCompile("^\\/[a-z0-9_-]*\\/?$")
)

// Handle WebRTC session description exchange via POST
func viewerPostHandler(w http.ResponseWriter, r *http.Request) {
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

	// Exchange session descriptions with WebRTC stream server
	remoteSdpChan <- remoteDescription
	localDescription := <-localSdpChan

	// Send server description as JSON
	jsonDesc, err := json.Marshal(localDescription)
	if err != nil {
		http.Error(w, "An error occurred while formating response", http.StatusInternalServerError)
		log.Println("An error occurred while sending session description", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonDesc)

	// Increment monitoring
	monitoring.WebSessions.Inc()
}

func viewerGetHandler(w http.ResponseWriter, r *http.Request) {
	// Render template
	data := struct {
		Path string
		Cfg  *Options
	}{Path: r.URL.Path[1:], Cfg: cfg}
	if err := templates.ExecuteTemplate(w, "base", data); err != nil {
		log.Println(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Increment monitoring
	monitoring.WebViewerServed.Inc()
}

// Handle site index and viewer pages
// POST requests are used to exchange WebRTC session descriptions
func viewerHandler(w http.ResponseWriter, r *http.Request) {
	// Validation on path
	if validPath.FindStringSubmatch(r.URL.Path) == nil {
		http.NotFound(w, r)
		log.Print(r.URL.Path)
		return
	}

	// Route depending on HTTP method
	switch r.Method {
	case http.MethodGet:
		viewerGetHandler(w, r)
	case http.MethodPost:
		viewerPostHandler(w, r)
	default:
		http.Error(w, "Sorry, only GET and POST methods are supported.", http.StatusBadRequest)
	}
}

// Handle static files
// We do not use http.FileServer as we do not want directory listing
func staticHandler(w http.ResponseWriter, r *http.Request) {
	path := "./web/" + r.URL.Path
	if f, err := os.Stat(path); err == nil && !f.IsDir() {
		http.ServeFile(w, r, path)
	} else {
		http.NotFound(w, r)
	}
}

// Load templates with pkger
// templates will be packed in the compiled binary
func loadTemplates() error {
	templates = template.New("")
	return pkger.Walk("/web/template", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip non-templates
		if info.IsDir() || !strings.HasSuffix(path, ".html") {
			return nil
		}

		// Open file with pkger
		f, err := pkger.Open(path)
		if err != nil {
			return err
		}

		// Read and parse template
		temp, err := ioutil.ReadAll(f)
		if err != nil {
			return err
		}
		templates, err = templates.Parse(string(temp))
		return err
	})
}

// Serve HTTP server
func Serve(rSdpChan chan webrtc.SessionDescription, lSdpChan chan webrtc.SessionDescription, c *Options) {
	remoteSdpChan = rSdpChan
	localSdpChan = lSdpChan
	cfg = c

	// Load templates
	if err := loadTemplates(); err != nil {
		log.Fatalln("Failed to load templates:", err)
	}

	// Set up HTTP router and server
	mux := http.NewServeMux()
	mux.HandleFunc("/", viewerHandler)
	mux.HandleFunc("/static/", staticHandler)
	log.Printf("HTTP server listening on %s", cfg.ListenAddress)
	log.Fatal(http.ListenAndServe(cfg.ListenAddress, mux))
}
