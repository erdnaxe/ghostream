package basic

import (
	"testing"
)

func TestBasicLogin(t *testing.T) {
	basicCredentials := make(map[string]string)
	basicCredentials["demo"] = "$2b$10$xuU7XFwmRX2CMgdSaA8rM.4Y8.BtRNzhUedwN0G8tCegDRNUERTCS"

	// Test good credentials
	backend, _ := New(&Options{Credentials: basicCredentials})
	ok, err := backend.Login("demo", "demo")
	if !ok {
		t.Error("Error while logging with the basic authentication:", err)
	}

	// Test bad username
	ok, err = backend.Login("baduser", "demo")
	if ok {
		t.Error("Authentification failed to fail:", err)
	}

	// Test bad password
	ok, err = backend.Login("demo", "badpass")
	if ok {
		t.Error("Authentification failed to fail:", err)
	}

	backend.Close()
}
