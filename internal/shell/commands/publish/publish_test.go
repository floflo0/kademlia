package publish_test

import (
	"bytes"
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"kademlia/internal/core/entities"
	"kademlia/internal/core/kademlia"
	"kademlia/internal/shell/commands/publish"
)

type mockDNS struct {
	pubKey ed25519.PublicKey
	err    error
}

func (m *mockDNS) GetPublicKey(domain string) (ed25519.PublicKey, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.pubKey, nil
}

func newMockKademlia(t *testing.T) kademlia.Kademlia {
	return kademlia.NewMockKademlia(
		func(firstContact *entities.Address) error {
			t.Fatal("Run should not be called")
			return nil
		},
		func(address entities.Address) (time.Duration, error) {
			t.Fatal("Ping should not be called")
			return 0, nil
		},
		func(value string) (*entities.KademliaID, error) {
			t.Fatal("Put should not be called")
			return nil, nil
		},
		func() []*kademlia.Bucket {
			t.Fatal("GetBucketsFunction should not be called")
			return []*kademlia.Bucket{}
		},
		func() []entities.KademliaID {
			t.Fatal("GetStoredKeysFunction should not be called")
			return []entities.KademliaID{}
		},
		func(string) (*string, *string, error) {
			t.Fatal("GetValueFunction should not be called")
			return nil, nil, nil
		},
		func() error {
			t.Fatal("Quit should not be called")
			return nil
		},
	)
}

func createTestFile(t *testing.T) string {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "binary.bin")
	if err := os.WriteFile(filePath, []byte("hello binary"), 0o644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	return filePath
}

func TestPublishCommand_Execute_MissingFlags(t *testing.T) {
	mockK := newMockKademlia(t)
	dns := &mockDNS{}
	var mockOut bytes.Buffer

	cmd := publish.NewPublishCommand(mockK, dns, &mockOut)
	err := cmd.Execute([]string{})
	if err == nil {
		t.Fatal("Expected error for missing required flags, got nil")
	}
}

func TestPublishCommand_Execute_InvalidKeyHex(t *testing.T) {
	filePath := createTestFile(t)
	mockK := newMockKademlia(t)
	dns := &mockDNS{}
	var mockOut bytes.Buffer

	cmd := publish.NewPublishCommand(mockK, dns, &mockOut)
	args := []string{"-k", "invalid_hex", "example.com:pkg:1.0.0", filePath}
	err := cmd.Execute(args)
	if err == nil {
		t.Fatal("Expected error for invalid key hex, got nil")
	}
}

func TestPublishCommand_Execute_Success(t *testing.T) {
	filePath := createTestFile(t)
	pubKey, privKey, _ := ed25519.GenerateKey(nil)
	privHex := hex.EncodeToString(privKey)

	dns := &mockDNS{pubKey: pubKey}
	mockK := newMockKademlia(t)

	mockKWithPublish := &mockKademliaWithPublish{
		Kademlia: mockK,
		publishFunc: func(domain, packageName, version, blob string, key ed25519.PrivateKey, dnsVerifier kademlia.DNSVerifier, force bool, explicitPrev string) (*kademlia.VersionRecord, error) {
			return &kademlia.VersionRecord{
				Tag:         "version-record",
				DomainName:  domain,
				PackageName: packageName,
				Version:     version,
				BlobHash:    "dummyhash",
			}, nil
		},
	}

	var mockOut bytes.Buffer
	cmd := publish.NewPublishCommand(mockKWithPublish, dns, &mockOut)

	args := []string{"-k", privHex, "example.com:my-lib:1.0.0", filePath}
	err := cmd.Execute(args)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !bytes.Contains(mockOut.Bytes(), []byte("Package successfully published!")) {
		t.Fatalf("Unexpected output: %s", mockOut.String())
	}
}

type mockKademliaWithPublish struct {
	kademlia.Kademlia
	publishFunc func(domain, packageName, version, blob string, privKey ed25519.PrivateKey, dns kademlia.DNSVerifier, force bool, explicitPrev string) (*kademlia.VersionRecord, error)
	installFunc func(domain, packageName, version string) (string, string, error)
}

func (m *mockKademliaWithPublish) PublishPackage(domain, packageName, version, blob string, privKey ed25519.PrivateKey, dns kademlia.DNSVerifier, force bool, explicitPrev string) (*kademlia.VersionRecord, error) {
	if m.publishFunc != nil {
		return m.publishFunc(domain, packageName, version, blob, privKey, dns, force, explicitPrev)
	}
	return nil, fmt.Errorf("not implemented")
}

func (m *mockKademliaWithPublish) InstallPackage(domain, packageName, version string) (string, string, error) {
	if m.installFunc != nil {
		return m.installFunc(domain, packageName, version)
	}
	return "", "", fmt.Errorf("not implemented")
}
