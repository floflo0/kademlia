package kademlia

import (
	"crypto/ed25519"
	"crypto/rand"
	"testing"
)

func TestValidateNewVersion_SuccessAndForkDetection(t *testing.T) {
	// 1. Setup keys and mock DNS
	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate keys: %v", err)
	}

	dns := NewMockDNSVerifier()
	dns.RegisterDomain("example.com", pubKey)

	// 2. Initial release v1.0.0
	v1 := &VersionRecord{
		Tag:                   "version-record",
		DomainName:            "example.com",
		PackageName:           "my-app",
		Version:               "1.0.0",
		BlobHash:              "hash_of_binary_v1",
		PrevVersionRecordHash: "",
	}
	v1.Sign(privKey)

	// Validate v1 (should pass)
	if err := ValidateNewVersion(v1, nil, pubKey); err != nil {
		t.Fatalf("failed to validate initial release: %v", err)
	}

	v1Hash, _ := v1.Hash()

	// 3. Second release v1.1.0 linked to v1
	v2 := &VersionRecord{
		Tag:                   "version-record",
		DomainName:            "example.com",
		PackageName:           "my-app",
		Version:               "1.1.0",
		BlobHash:              "hash_of_binary_v2",
		PrevVersionRecordHash: v1Hash,
	}
	v2.Sign(privKey)

	// Validate v2 (should pass)
	if err := ValidateNewVersion(v2, v1, pubKey); err != nil {
		t.Fatalf("failed to validate valid update: %v", err)
	}

	// 4. Fork attempt: v2Fork tries to update from v1 with different content
	v2Fork := &VersionRecord{
		Tag:                   "version-record",
		DomainName:            "example.com",
		PackageName:           "my-app",
		Version:               "1.1.0",
		BlobHash:              "hash_of_malicious_binary",
		PrevVersionRecordHash: "wrong_previous_hash",
	}
	v2Fork.Sign(privKey)

	// Validate fork attempt against v1 (should fail with ErrForkDetected)
	if err := ValidateNewVersion(v2Fork, v1, pubKey); err != ErrForkDetected {
		t.Fatalf("expected ErrForkDetected, got: %v", err)
	}
}
