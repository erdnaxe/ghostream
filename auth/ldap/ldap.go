package ldap

import (
	"github.com/go-ldap/ldap/v3"
)

// Options holds package configuration
type Options struct {
	URI    string
	UserDn string
}

// LDAP authentification backend
type LDAP struct {
	Cfg Options
}

// Login tries to bind to LDAP
// Returns (true, nil) if success
func (a LDAP) Login(username string, password string) (bool, error) {
	// Connect to LDAP server
	l, err := ldap.DialURL(a.Cfg.URI)
	if err != nil {
		return false, err
	}
	defer l.Close()

	// Try to bind as user
	err = l.Bind("cn=username,dc=example,dc=com", password)
	if err != nil {
		return false, err
	}

	// Login succeeded
	return true, nil
}
