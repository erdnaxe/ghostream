package monitoring

import (
	"net/http"
	"testing"
	"time"
)

func TestMonitoringServe(t *testing.T) {
	go Serve(&Options{Enabled: false, ListenAddress: "127.0.0.1:2112"})

	// Sleep 0.5 second to ensure that the web server is running, to avoid fails because the request came too early
	time.Sleep(500 * time.Millisecond)

	// Test GET request, should fail
	resp, err := http.Get("http://127.0.0.1:2112/metrics")
	if err == nil {
		t.Error("Failed to fail task")
	}

	go Serve(&Options{Enabled: true, ListenAddress: "127.0.0.1:2112"})

	// Sleep 0.5 second to ensure that the web server is running, to avoid fails because the request came too early
	time.Sleep(500 * time.Millisecond)

	// Test GET request
	resp, err = http.Get("http://127.0.0.1:2112/metrics")
	if err != nil {
		t.Error("Error while getting /metrics:", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Metric page returned %v != %v on GET", resp.StatusCode, http.StatusOK)
	}
}
