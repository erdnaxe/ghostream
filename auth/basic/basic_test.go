package basic

import (
	"testing"
)

func TestBasicLogin(t *testing.T) {
	basicCredentials := make(map[string]string)
	basicCredentials["demo"] = "$2b$15$LRnG3eIHFlYIguTxZOLH7eHwbQC/vqjnLq6nDFiHSUDKIU.f5/1H6"

	// Test good credentials
	backend, _ := New(&Options{Credentials: basicCredentials})
	ok, err := backend.Login("demo", "demo")
	if !ok {
		t.Error("Error while logging with the basic authentication:", err)
	}

	// Test bad credentials
	ok, err = backend.Login("demo", "badpass")
	if ok {
		t.Error("Authentification failed to fail:", err)
	}

	backend.Close()
}
