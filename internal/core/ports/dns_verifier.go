package ports

import (
	"crypto/ed25519"
	"errors"
)

type DNSVerifier interface {
	GetPublicKey(domain string) (ed25519.PublicKey, error)
}

var ErrDomainNotFound = errors.New(
	"domain TXT record not found or public key missing",
)
