// Package ldap provides a LDAP authentification backend
package ldap

import (
	"github.com/go-ldap/ldap/v3"
	"log"
)

// Options holds package configuration
type Options struct {
	Aliases map[string]string
	URI     string
	UserDn  string
}

// LDAP authentification backend
type LDAP struct {
	Cfg  *Options
	Conn *ldap.Conn
}

// Login tries to bind to LDAP
// Returns (true, nil) if success
func (a LDAP) Login(username string, password string) (bool, error) {
	// Resolve stream alias if necessary
	for aliasFor, ok := a.Cfg.Aliases[username]; ok; aliasFor, ok = a.Cfg.Aliases[username] {
		log.Printf("[LDAP] Use stream alias %s for username %s", username, aliasFor)
		username = aliasFor
	}

	// Try to bind as user
	bindDn := "cn=" + username + "," + a.Cfg.UserDn
	err := a.Conn.Bind(bindDn, password)

	// Login succeeded if no error
	return err == nil, err
}

// Close LDAP connection
func (a LDAP) Close() {
	a.Conn.Close()
}

// New instanciates a new LDAP connection
func New(cfg *Options) (LDAP, error) {
	backend := LDAP{Cfg: cfg}

	// Connect to LDAP server
	c, err := ldap.DialURL(backend.Cfg.URI)
	backend.Conn = c
	return backend, err
}
