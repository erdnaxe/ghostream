// Package basic provides a basic authentification backend
package basic

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

// To generate bcrypt hashed password from Python,
// python3 -c 'import bcrypt; print(bcrypt.hashpw(b"PASSWORD", bcrypt.gensalt(rounds=15)).decode("ascii"))'

// Options holds package configuration
type Options struct {
	// Username: hashedPassword
	Credentials map[string]string
}

// Basic authentification backend
type Basic struct {
	Cfg *Options
}

// Login hashs password and compare
// Returns (true, nil) if success
func (a Basic) Login(username string, password string) (bool, error) {
	hash, ok := a.Cfg.Credentials[username]
	if !ok {
		return false, errors.New("user not found in credentials")
	}
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))

	// Login succeeded if no error
	return err == nil, err
}

// Close has no connection to close
func (a Basic) Close() {
}

// New instanciates a new Basic authentification backend
func New(cfg *Options) (Basic, error) {
	backend := Basic{Cfg: cfg}
	return backend, nil
}
