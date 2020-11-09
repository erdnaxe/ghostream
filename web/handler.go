// Package web serves the JavaScript player and WebRTC negotiation
package web

import (
	"bytes"
	"encoding/json"
	"github.com/markbates/pkger"
	"gitlab.crans.org/nounous/ghostream/internal/monitoring"
	"gitlab.crans.org/nounous/ghostream/stream/ovenmediaengine"
	"gitlab.crans.org/nounous/ghostream/stream/webrtc"
	"html/template"
	"log"
	"net"
	"net/http"
	"regexp"
	"strings"
	"time"
)

var (
	// Precompile regex
	validPath = regexp.MustCompile("^/[a-z0-9@_-]*$")

	connectedClients = make(map[string]map[string]int64)
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
	// Retrieve stream name from URL
	name := strings.SplitN(strings.Replace(r.URL.Path[7:], "/", "", -1), "@", 2)[0]
	userCount := 0

	// Clients have a unique generated identifier per session, that expires in 40 seconds.
	// Each time the client connects to this page, the identifier is renewed.
	// Yeah, that's not a good way to have stats, but it works...
	if connectedClients[name] == nil {
		connectedClients[name] = make(map[string]int64)
	}
	currentTime := time.Now().Unix()
	if _, ok := r.URL.Query()["uid"]; ok {
		uid := r.URL.Query()["uid"][0]
		connectedClients[name][uid] = currentTime
	}
	toDelete := make([]string, 0)
	for uid, oldTime := range connectedClients[name] {
		if currentTime-oldTime > 40 {
			toDelete = append(toDelete, uid)
		}
	}
	for _, uid := range toDelete {
		delete(connectedClients[name], uid)
	}

	// Get requested stream
	stream, err := streams.Get(name)
	if err == nil {
		userCount = stream.ClientCount()
		userCount += webrtc.GetNumberConnectedSessions(name)
		userCount += len(connectedClients[name])
	}

	// Display connected users statistics
	enc := json.NewEncoder(w)
	err = enc.Encode(struct{ ConnectedViewers int }{userCount})
	if err != nil {
		http.Error(w, "Failed to generate JSON.", http.StatusInternalServerError)
		log.Printf("Failed to generate JSON: %s", err)
	}
}
