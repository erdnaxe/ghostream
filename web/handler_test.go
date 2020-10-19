package web

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"gitlab.crans.org/nounous/ghostream/messaging"
)

func TestViewerPageGET(t *testing.T) {
	// Load templates
	if err := loadTemplates(); err != nil {
		t.Errorf("Failed to load templates: %v", err)
	}

	// Init streams messaging
	streams = messaging.New()

	cfg = &Options{}

	// Test GET request
	r, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	http.HandlerFunc(viewerHandler).ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Errorf("Viewer page returned %v != %v on GET", w.Code, http.StatusOK)
	}

	// Test GET request on not found page
	r, _ = http.NewRequest("GET", "", nil)
	w = httptest.NewRecorder()
	http.HandlerFunc(viewerHandler).ServeHTTP(w, r)
	if w.Code != http.StatusNotFound {
		t.Errorf("Viewer page returned %v != %v on GET", w.Code, http.StatusOK)
	}

	// Test GET request on statistics page
	r, _ = http.NewRequest("GET", "/_stats/demo/", nil)
	w = httptest.NewRecorder()
	http.HandlerFunc(statisticsHandler).ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Errorf("Viewer page returned %v != %v on GET", w.Code, http.StatusOK)
	}
}
