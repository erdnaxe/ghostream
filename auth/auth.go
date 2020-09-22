package auth

import (
	"gitlab.crans.org/nounous/ghostream/auth/ldap"
)

// Options holds web package configuration
type Options struct {
	Backend string
	LDAP    ldap.Options
}
