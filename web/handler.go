package web

import (
	"bytes"
	"encoding/json"
	"html/template"
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/markbates/pkger"
	"gitlab.crans.org/nounous/ghostream/internal/monitoring"
	"gitlab.crans.org/nounous/ghostream/stream/webrtc"
)

// Handle WebRTC session description exchange via POST
func viewerPostHandler(w http.ResponseWriter, r *http.Request) {
	// Limit response body to 128KB
	r.Body = http.MaxBytesReader(w, r.Body, 131072)

	// Get stream ID from URL, or from domain name
	path := r.URL.Path[1:]
	if cfg.OneStreamPerDomain {
		host := r.Host
		if strings.Contains(host, ":") {
			realHost, _, err := net.SplitHostPort(r.Host)
			if err != nil {
				log.Printf("Failed to split host and port from %s", r.Host)
				return
			}
			host = realHost
		}
		path = host
	}

	// Decode client description
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	remoteDescription := webrtc.SessionDescription{}
	if err := dec.Decode(&remoteDescription); err != nil {
		http.Error(w, "The JSON WebRTC offer is malformed", http.StatusBadRequest)
		return
	}

	// Exchange session descriptions with WebRTC stream server
	remoteSdpChan <- struct {
		StreamID          string
		RemoteDescription webrtc.SessionDescription
	}{StreamID: path, RemoteDescription: remoteDescription}
	localDescription := <-localSdpChan

	// Send server description as JSON
	jsonDesc, err := json.Marshal(localDescription)
	if err != nil {
		http.Error(w, "An error occurred while formating response", http.StatusInternalServerError)
		log.Println("An error occurred while sending session description", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(jsonDesc)
	if err != nil {
		log.Println("An error occurred while sending session description", err)
	}

	// Increment monitoring
	monitoring.WebSessions.Inc()
}

func viewerGetHandler(w http.ResponseWriter, r *http.Request) {
	// Get stream ID from URL, or from domain name
	path := r.URL.Path[1:]
	if cfg.OneStreamPerDomain {
		host := r.Host
		if strings.Contains(host, ":") {
			realHost, _, err := net.SplitHostPort(r.Host)
			if err != nil {
				log.Printf("Failed to split host and port from %s", r.Host)
				return
			}
			host = realHost
		}
		path = host
	}

	// Render template
	data := struct {
		Cfg       *Options
		Path      string
		WidgetURL string
	}{Path: path, Cfg: cfg}

	// Compute the WidgetURL with the stream path
	b := &bytes.Buffer{}
	_ = template.Must(template.New("").Parse(cfg.WidgetURL)).Execute(b, data)
	data.WidgetURL = b.String()

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

func staticHandler() http.Handler {
	// Set up static files server
	staticFs := http.FileServer(pkger.Dir("/web/static"))
	return http.StripPrefix("/static/", staticFs)
}

func statisticsHandler(w http.ResponseWriter, r *http.Request) {
	// Display connected users stats
	enc := json.NewEncoder(w)
	err := enc.Encode(struct {
		ConnectedViewers int
	}{webrtc.GetNumberConnectedSessions()})
	if err != nil {
		http.Error(w, "Failed to generate JSON.", http.StatusInternalServerError)
		log.Printf("Failed to generate JSON: %s", err)
	}
}
