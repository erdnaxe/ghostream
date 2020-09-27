package web

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestViewerPageGET(t *testing.T) {
	r, _ := http.NewRequest("GET", "", nil)
	w := httptest.NewRecorder()
	http.HandlerFunc(viewerHandler).ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Errorf("Viewer page didn't return %v on GET", http.StatusOK)
	}
}
