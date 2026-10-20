package shell

import (
	"crypto/ed25519"
	"kademlia/internal/core/entities"
	"kademlia/internal/core/kademlia"
	"kademlia/internal/core/ports"
	"testing"
	"time"
)

type testDNS struct{}

func (m *testDNS) GetPublicKey(domain string) (ed25519.PublicKey, error) {
	return nil, nil
}

func TestShellCompleter(t *testing.T) {
	mockK := kademlia.NewMockKademlia(
		func(*entities.Address) error { return nil },
		func(entities.Address) (time.Duration, error) { return 0, nil },
		func(string) (*entities.KademliaID, error) { return nil, nil },
		func() []*kademlia.Bucket { return nil },
		func() []entities.KademliaID { return nil },
		func(string) (*string, *string, error) { return nil, nil, nil },
		func() error { return nil },
		func(string, string, string, string, ed25519.PrivateKey, ports.DNSVerifier, bool, string) (*kademlia.VersionRecord, error) {
			return nil, nil
		},
		func(string, string, string) (string, string, error) { return "", "", nil },
		func(string, string) ([]*kademlia.VersionRecord, error) { return nil, nil },
	)
	s := NewShell(mockK, &testDNS{})
	c := s.completer()
	if c == nil {
		t.Fatalf("expected completer, got nil")
	}
}
