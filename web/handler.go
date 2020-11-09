// Package web serves the JavaScript player and WebRTC negotiation
package web

import (
	"bytes"
	"encoding/json"
	"html/template"
	"log"
	"net"
	"net/http"
	"regexp"
	"strings"

	"github.com/markbates/pkger"
	"gitlab.crans.org/nounous/ghostream/internal/monitoring"
	"gitlab.crans.org/nounous/ghostream/stream/ovenmediaengine"
	"gitlab.crans.org/nounous/ghostream/stream/webrtc"
)

var (
	// Precompile regex
	validPath = regexp.MustCompile("^/[a-z0-9@_-]*$")
)

// Handle site index and viewer pages
func viewerHandler(w http.ResponseWriter, r *http.Request) {
	// Validation on path
	if validPath.FindStringSubmatch(r.URL.Path) == nil {
		http.NotFound(w, r)
		log.Printf("Replied not found on %s", r.URL.Path)
		return
	}

	// Check method
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed.", http.StatusMethodNotAllowed)
	}

	// Get stream ID from URL, or from domain name
	path := r.URL.Path[1:]
	host := r.Host
	if strings.Contains(host, ":") {
		realHost, _, err := net.SplitHostPort(r.Host)
		if err != nil {
			log.Printf("Failed to split host and port from %s", r.Host)
			return
		}
		host = realHost
	}
	host = strings.Replace(host, ".", "-", -1)
	if streamID, ok := cfg.MapDomainToStream[host]; ok {
		// Move home page to /about
		if path == "about" {
			path = ""
		} else {
			path = streamID
		}
	}

	// Render template
	data := struct {
		Cfg       *Options
		Path      string
		WidgetURL string
		OMECfg    *ovenmediaengine.Options
	}{Path: path, Cfg: cfg, WidgetURL: "", OMECfg: omeCfg}

	// Load widget is user does not disable it with ?nowidget
	if _, ok := r.URL.Query()["nowidget"]; !ok {
		// Compute the WidgetURL with the stream path
		b := &bytes.Buffer{}
		_ = template.Must(template.New("").Parse(cfg.WidgetURL)).Execute(b, data)
		data.WidgetURL = b.String()
	}

	if err := templates.ExecuteTemplate(w, "base", data); err != nil {
		log.Println(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Increment monitoring
	monitoring.WebViewerServed.Inc()
}

func staticHandler() http.Handler {
	// Set up static files server
	staticFs := http.FileServer(pkger.Dir("/web/static"))
	return http.StripPrefix("/static/", staticFs)
}

func statisticsHandler(w http.ResponseWriter, r *http.Request) {
	name := strings.SplitN(strings.Replace(r.URL.Path[7:], "/", "", -1), "@", 2)[0]
	userCount := 0

	// Get requested stream
	stream, err := streams.Get(name)
	if err == nil {
		userCount = stream.ClientCount()
		userCount += webrtc.GetNumberConnectedSessions(name)
	}

	// Display connected users statistics
	enc := json.NewEncoder(w)
	err = enc.Encode(struct{ ConnectedViewers int }{userCount})
	if err != nil {
		http.Error(w, "Failed to generate JSON.", http.StatusInternalServerError)
		log.Printf("Failed to generate JSON: %s", err)
	}
}
