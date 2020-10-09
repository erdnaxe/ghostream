package auth

import (
	"testing"

	"gitlab.crans.org/nounous/ghostream/auth/basic"
	"gitlab.crans.org/nounous/ghostream/auth/ldap"
)

func TestLoadConfiguration(t *testing.T) {
	// Test to create a basic authentification backend
	_, err := New(&Options{
		Enabled: true,
		Backend: "basic",
		Basic:   basic.Options{Credentials: make(map[string]string)},
	})
	if err != nil {
		t.Error("Creating basic authentication backend failed:", err)
	}

	// Test to create a LDAP authentification backend
	// FIXME Should fail as there is no LDAP server
	_, err = New(&Options{
		Enabled: true,
		Backend: "ldap",
		LDAP:    ldap.Options{URI: "ldap://127.0.0.1:389", UserDn: "cn=users,dc=example,dc=com"},
	})
	if err == nil {
		t.Error("Creating ldap authentication backend successed mysteriously:", err)
	}

	// Test to bypass authentification backend
	backend, err := New(&Options{
		Enabled: false,
	})
	if backend != nil {
		t.Error("Failed to bypass authentication backend:", err)
	}
}
