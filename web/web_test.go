package web

import (
	"net/http"
	"testing"
	"time"
)

// TestHTTPServe tries to serve a real HTTP server and load some pages
func TestHTTPServe(t *testing.T) {
	// Load templates
	if err := loadTemplates(); err != nil {
		t.Errorf("Failed to load templates: %v", err)
	}

	go Serve(nil, nil, &Options{ListenAddress: "127.0.0.1:8081"})

	// Sleep 1 second to ensure that the web server is running, to avoid fails because the request came too early
	time.Sleep(1000000000)

	// Test GET request
	resp, err := http.Get("http://localhost:8081/")
	if err != nil {
		t.Error("Error while getting /:", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Viewer page returned %v != %v on GET", resp.StatusCode, http.StatusOK)
	}

	resp, err = http.Get("http://localhost:8081/demo")
	if err != nil {
		t.Error("Error while getting /demo:", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Viewer page returned %v != %v on GET", resp.StatusCode, http.StatusOK)
	}

	resp, err = http.Get("http://localhost:8081/static/js/viewer.js")
	if err != nil {
		t.Error("Error while getting /static/js/viewer/js:", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Viewer page returned %v != %v on GET", resp.StatusCode, http.StatusOK)
	}

	resp, err = http.Get("http://localhost:8081/_stats/demo/")
	if err != nil {
		t.Error("Error while getting /_stats:", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Viewer page returned %v != %v on GET", resp.StatusCode, http.StatusOK)
	}
}
