package auth

import (
	"testing"

	"gitlab.crans.org/nounous/ghostream/auth/basic"
	"gitlab.crans.org/nounous/ghostream/auth/ldap"
)

func TestLoadConfiguration(t *testing.T) {
	_, err := New(&Options{
		Enabled: true,
		Backend: "basic",
		Basic:   basic.Options{Credentials: make(map[string]string)},
	})

	if err != nil {
		t.Error("Creating basic authentication backend failed:", err)
	}

	_, err = New(&Options{
		Enabled: true,
		Backend: "ldap",
		LDAP:    ldap.Options{URI: "ldap://127.0.0.1:389", UserDn: "cn=users,dc=example,dc=com"},
	})

	// TODO Maybe start a LDAP server?
	if err == nil {
		t.Error("Creating ldap authentication backend successed mysteriously:", err)
	}
}
