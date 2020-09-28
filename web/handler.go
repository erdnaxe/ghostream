package web

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/markbates/pkger"
	"github.com/pion/webrtc/v3"
	"gitlab.crans.org/nounous/ghostream/internal/monitoring"
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

func staticHandler() http.Handler {
	// Set up static files server
	staticFs := http.FileServer(pkger.Dir("/web/static"))
	return http.StripPrefix("/static/", staticFs)
}
