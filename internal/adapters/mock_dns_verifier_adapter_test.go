package adapters_test

import (
	"crypto/ed25519"
	"errors"
	"testing"

	"kademlia/internal/adapters"
	"kademlia/internal/core/ports"
)

func newKey(t *testing.T) ed25519.PublicKey {
	t.Helper()
	pub, _, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	return pub
}

func TestMockDNSVerifierAdapter_GetPublicKey_UnknownDomain(t *testing.T) {
	mockDNSVerifier := adapters.NewMockDNSVerifierAdapter()

	key, err := mockDNSVerifier.GetPublicKey("unknown.example")
	if !errors.Is(err, ports.ErrDomainNotFound) {
		t.Fatalf("expected ErrDomainNotFound, got %v", err)
	}
	if key != nil {
		t.Fatalf("expected nil key, got %x", key)
	}
}

func TestMockDNSVerifierAdapter_RegisterThenGet(t *testing.T) {
	mockDNSVerifier := adapters.NewMockDNSVerifierAdapter()
	want := newKey(t)

	mockDNSVerifier.RegisterDomain("example.com", want)

	got, err := mockDNSVerifier.GetPublicKey("example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got.Equal(want) {
		t.Fatalf("key mismatch: got %x, want %x", got, want)
	}
}

func TestMockDNSVerifierAdapter_RegisterOverwrites(t *testing.T) {
	mockDNSVerifier := adapters.NewMockDNSVerifierAdapter()
	first, second := newKey(t), newKey(t)

	mockDNSVerifier.RegisterDomain("example.com", first)
	mockDNSVerifier.RegisterDomain("example.com", second)

	got, err := mockDNSVerifier.GetPublicKey("example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got.Equal(second) {
		t.Fatal("expected second key to replace the first")
	}
}

func TestMockDNSVerifierAdapter_DomainsAreIndependent(t *testing.T) {
	mockDNSVerifier := adapters.NewMockDNSVerifierAdapter()
	keyA, keyB := newKey(t), newKey(t)

	mockDNSVerifier.RegisterDomain("a.example", keyA)
	mockDNSVerifier.RegisterDomain("b.example", keyB)

	tests := []struct {
		domain string
		want   ed25519.PublicKey
	}{
		{"a.example", keyA},
		{"b.example", keyB},
	}
	for _, tt := range tests {
		t.Run(tt.domain, func(t *testing.T) {
			got, err := mockDNSVerifier.GetPublicKey(tt.domain)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !got.Equal(tt.want) {
				t.Fatalf("key mismatch for %s", tt.domain)
			}
		})
	}
}

func TestMockDNSVerifierAdapter_DomainLookupIsExact(t *testing.T) {
	mockDNSVerifier := adapters.NewMockDNSVerifierAdapter()
	mockDNSVerifier.RegisterDomain("example.com", newKey(t))

	for _, domain := range []string{"", "EXAMPLE.COM", "sub.example.com", "example.com."} {
		t.Run(domain, func(t *testing.T) {
			if _, err := mockDNSVerifier.GetPublicKey(domain); !errors.Is(err, ports.ErrDomainNotFound) {
				t.Fatalf("expected ErrDomainNotFound for %q, got %v", domain, err)
			}
		})
	}
}
