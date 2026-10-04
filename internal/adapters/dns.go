package adapters

import (
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"net"
	"strings"
)

type DNSAdapter struct{}

func NewDNSAdapter() *DNSAdapter {
	return &DNSAdapter{}
}

func (d *DNSAdapter) GetPublicKey(domain string) (ed25519.PublicKey, error) {
	txts, err := net.LookupTXT(domain)
	if err != nil {
		return nil, fmt.Errorf("failed DNS lookup for %s: %w", domain, err)
	}
	for _, txt := range txts {
		pubBytes, err := hex.DecodeString(strings.TrimSpace(txt))
		if err == nil && len(pubBytes) == ed25519.PublicKeySize {
			return ed25519.PublicKey(pubBytes), nil
		}
	}
	return nil, fmt.Errorf("no valid public key found in TXT records for %s", domain)
}
