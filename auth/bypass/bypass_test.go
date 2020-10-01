package bypass

import (
	"testing"
)

func TestBypassLogin(t *testing.T) {
	backend, _ := New()
	ok, err := backend.Login("demo", "demo")
	if !ok {
		t.Error("Error while logging with the bypass authentication:", err)
	}
	backend.Close()
}
