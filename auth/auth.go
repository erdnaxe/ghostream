package auth

import (
	"errors"
	"log"
	"strings"

	"gitlab.crans.org/nounous/ghostream/auth/basic"
	"gitlab.crans.org/nounous/ghostream/auth/bypass"
	"gitlab.crans.org/nounous/ghostream/auth/ldap"
)

// Options holds package configuration
type Options struct {
	Backend string
	Basic   basic.Options
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

	switch strings.ToLower(cfg.Backend) {
	case "basic":
		backend, err = basic.New(&cfg.Basic)
	case "bypass":
		backend, err = bypass.New()
	case "ldap":
		backend, err = ldap.New(&cfg.LDAP)
	default:
		// Package is misconfigured
		backend, err = nil, errors.New("authentification backend not found")
	}

	if err != nil {
		// Backend init failed
		return nil, err
	}

	log.Printf("%s backend successfully initialized", cfg.Backend)
	return backend, nil
}
