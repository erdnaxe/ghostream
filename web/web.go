// Package web serves the JavaScript player and WebRTC negociation
package web

import (
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/markbates/pkger"
	"github.com/pion/webrtc/v3"
)

// Options holds web package configuration
type Options struct {
	Enabled                     bool
	CustomCSS                   string
	Favicon                     string
	Hostname                    string
	ListenAddress               string
	Name                        string
	MapDomainToStream           map[string]string
	PlayerPoster                string
	SRTServerPort               string
	STUNServers                 []string
	ViewersCounterRefreshPeriod int
	WidgetURL                   string
}

var (
	cfg *Options

	// WebRTC session description channels
	remoteSdpChan chan struct {
		StreamID          string
		RemoteDescription webrtc.SessionDescription
	}
	localSdpChan chan webrtc.SessionDescription

	// Preload templates
	templates *template.Template

	// Precompile regex
	validPath = regexp.MustCompile("^/[a-z0-9@_\\-]*/?$")
)

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
func Serve(rSdpChan chan struct {
	StreamID          string
	RemoteDescription webrtc.SessionDescription
}, lSdpChan chan webrtc.SessionDescription, c *Options) {
	remoteSdpChan = rSdpChan
	localSdpChan = lSdpChan
	cfg = c

	if !cfg.Enabled {
		// Web server is not enabled, ignore
		return
	}

	// Load templates
	if err := loadTemplates(); err != nil {
		log.Fatalln("Failed to load templates:", err)
	}

	// Set up HTTP router and server
	mux := http.NewServeMux()
	mux.HandleFunc("/", viewerHandler)
	mux.Handle("/static/", staticHandler())
	mux.HandleFunc("/_stats/", statisticsHandler)
	log.Printf("HTTP server listening on %s", cfg.ListenAddress)
	log.Fatal(http.ListenAndServe(cfg.ListenAddress, mux))
}
