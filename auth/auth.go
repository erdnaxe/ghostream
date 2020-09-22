package auth

import (
	"errors"

	"gitlab.crans.org/nounous/ghostream/auth/ldap"
)

// Options holds package configuration
type Options struct {
	Backend string
	LDAP    ldap.Options
}

// Backend to log user in
type Backend interface {
	Login(string, string) (bool, error)
	Close()
}

// New initialize authentification backend
func New(cfg *Options) (Backend, error) {
	var backend Backend
	var err error

	if cfg.Backend == "LDAP" {
		backend, err = ldap.NewLDAP(&cfg.LDAP)
		if err != nil {
			return nil, err
		}
	} else {
		// Package is misconfigured
		return nil, errors.New("Authentification backend not found")
	}

	// Init and return backend
	return backend, nil
}
