package web

import (
	"net/http"
	"testing"
)

// TestHTTPServe tries to serve a real HTTP server and load some pages
func TestHTTPServe(t *testing.T) {
	// Load templates
	if err := loadTemplates(); err != nil {
		t.Errorf("Failed to load templates: %v", err)
	}

	go Serve(nil, nil, &Options{ListenAddress: "127.0.0.1:8081"})
	// Test GET request
	resp, _ := http.Get("http://localhost:8081/")
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Viewer page returned %v != %v on GET", resp.StatusCode, http.StatusOK)
	}
	resp, _ = http.Get("http://localhost:8081/demo")
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Viewer page returned %v != %v on GET", resp.StatusCode, http.StatusOK)
	}
	resp, _ = http.Get("http://localhost:8081/static/js/viewer.js")
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Viewer page returned %v != %v on GET", resp.StatusCode, http.StatusOK)
	}
	resp, _ = http.Get("http://localhost:8081/_stats")
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Viewer page returned %v != %v on GET", resp.StatusCode, http.StatusOK)
	}
}
