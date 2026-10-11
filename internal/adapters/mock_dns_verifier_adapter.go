package adapters

import (
	"crypto/ed25519"
	"kademlia/internal/core/ports"
)

// MockDNSVerifierAdapter simulates a DNS server for tests in memory
type MockDNSVerifierAdapter struct {
	records map[string]ed25519.PublicKey
}

// NewMockDNSVerifierAdapter instance a simulates DNS verifier
func NewMockDNSVerifierAdapter() *MockDNSVerifierAdapter {
	return &MockDNSVerifierAdapter{
		records: make(map[string]ed25519.PublicKey),
	}
}

// RegisterDomain asociates a domain name to a public key
func (m *MockDNSVerifierAdapter) RegisterDomain(domain string, pubKey ed25519.PublicKey) {
	m.records[domain] = pubKey
}

// GetPublicKey returns the public key asociated to the domain
func (m *MockDNSVerifierAdapter) GetPublicKey(domain string) (ed25519.PublicKey, error) {
	pubKey, exists := m.records[domain]
	if !exists {
		return nil, ports.ErrDomainNotFound
	}
	return pubKey, nil
}
