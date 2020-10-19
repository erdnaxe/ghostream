package web

import (
	"net/http"
	"testing"
	"time"

	"gitlab.crans.org/nounous/ghostream/messaging"
)

// TestHTTPServe tries to serve a real HTTP server and load some pages
func TestHTTPServe(t *testing.T) {
	// Init streams messaging
	streams := messaging.New()

	// Create a disabled web server
	go Serve(streams, nil, nil, &Options{Enabled: false, ListenAddress: "127.0.0.1:8081"})

	// Sleep 500ms to ensure that the web server is running, to avoid fails because the request came too early
	time.Sleep(500 * time.Millisecond)

	// Test GET request, should fail
	resp, err := http.Get("http://localhost:8081/")
	if err == nil {
		t.Error("Web server did init with Enabled=false")
	}

	// Now let's really start the web server
	go Serve(streams, nil, nil, &Options{Enabled: true, ListenAddress: "127.0.0.1:8081"})

	// Sleep 500ms to ensure that the web server is running, to avoid fails because the request came too early
	time.Sleep(500 * time.Millisecond)

	// Test home page
	resp, err = http.Get("http://localhost:8081/")
	if err != nil {
		t.Errorf("Error while getting /: %s", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Viewer page returned %v != %v on GET", resp.StatusCode, http.StatusOK)
	}

	// Test viewer page
	resp, err = http.Get("http://localhost:8081/demo")
	if err != nil {
		t.Errorf("Error while getting /demo: %s", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Viewer page returned %v != %v on GET", resp.StatusCode, http.StatusOK)
	}

	// Test viewer static file
	resp, err = http.Get("http://localhost:8081/static/js/viewer.js")
	if err != nil {
		t.Errorf("Error while getting /static/js/viewer/js: %s", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Viewer page returned %v != %v on GET", resp.StatusCode, http.StatusOK)
	}

	// Test viewer statistics endpoint
	resp, err = http.Get("http://localhost:8081/_stats/demo/")
	if err != nil {
		t.Errorf("Error while getting /_stats: %s", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Viewer page returned %v != %v on GET", resp.StatusCode, http.StatusOK)
	}
}
