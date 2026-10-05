package kademlia

import (
	"crypto/ed25519"
	"errors"
)

var ErrDomainNotFound = errors.New("domain TXT record not found or public key missing")

// DNSVerifier interface to verify the identity of the owner developer of the domain
type DNSVerifier interface {
	GetPublicKey(domain string) (ed25519.PublicKey, error)
}

// MockDNSVerifier simulates a DNS server for tests in memory
type MockDNSVerifier struct {
	records map[string]ed25519.PublicKey
}

// NewMockDNSVerifier instance a simulates DNS verifier
func NewMockDNSVerifier() *MockDNSVerifier {
	return &MockDNSVerifier{
		records: make(map[string]ed25519.PublicKey),
	}
}

// RegisterDomain asociates a domain name to a public key
func (m *MockDNSVerifier) RegisterDomain(domain string, pubKey ed25519.PublicKey) {
	m.records[domain] = pubKey
}

// GetPublicKey returns the public key asociated to the domain
func (m *MockDNSVerifier) GetPublicKey(domain string) (ed25519.PublicKey, error) {
	pubKey, exists := m.records[domain]
	if !exists {
		return nil, ErrDomainNotFound
	}
	return pubKey, nil
}
