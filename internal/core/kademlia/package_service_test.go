package kademlia

import (
	"crypto/ed25519"
	"crypto/rand"
	"kademlia/internal/adapters"
	"testing"
)

func newTestKademliaNode() *kademlia {
	return &kademlia{
		dataStore: adapters.NewInMemoryDataStore(),
	}
}

func TestPackageService_PublishAndInstall(t *testing.T) {
	k := newTestKademliaNode()

	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate keys: %v", err)
	}

	dns := NewMockDNSVerifier()
	dns.RegisterDomain("example.com", pubKey)

	domain := "example.com"
	pkg := "my-lib"

	// 1. Publish initial release v1.0.0
	blobV1 := "binary_content_v1.0.0"
	_, err = k.PublishPackage(domain, pkg, "1.0.0", blobV1, privKey, dns)
	if err != nil {
		t.Fatalf("failed to publish v1.0.0: %v", err)
	}

	// 2. Install latest (should resolve v1.0.0)
	installedBlob, resolvedVer, err := k.InstallPackage(domain, pkg, "latest")
	if err != nil {
		t.Fatalf("failed to install latest package: %v", err)
	}
	if installedBlob != blobV1 || resolvedVer != "1.0.0" {
		t.Fatalf("expected v1.0.0 content, got ver=%s blob=%s", resolvedVer, installedBlob)
	}

	// 3. Publish update v1.1.0
	blobV2 := "binary_content_v1.1.0"
	_, err = k.PublishPackage(domain, pkg, "1.1.0", blobV2, privKey, dns)
	if err != nil {
		t.Fatalf("failed to publish v1.1.0: %v", err)
	}

	// 4. Install latest again (should now resolve v1.1.0)
	installedBlob, resolvedVer, err = k.InstallPackage(domain, pkg, "")
	if err != nil {
		t.Fatalf("failed to install updated package: %v", err)
	}
	if installedBlob != blobV2 || resolvedVer != "1.1.0" {
		t.Fatalf("expected v1.1.0 content, got ver=%s blob=%s", resolvedVer, installedBlob)
	}

	// 5. Install explicit older version v1.0.0 via history lookup
	installedBlob, resolvedVer, err = k.InstallPackage(domain, pkg, "1.0.0")
	if err != nil {
		t.Fatalf("failed to install historical v1.0.0: %v", err)
	}
	if installedBlob != blobV1 || resolvedVer != "1.0.0" {
		t.Fatalf("expected historical v1.0.0, got ver=%s blob=%s", resolvedVer, installedBlob)
	}

	// 6. Attempt rollback (v1.0.5 < v1.1.0) -> should fail with validation error
	_, err = k.PublishPackage(domain, pkg, "1.0.5", "invalid_rollback", privKey, dns)
	if err == nil {
		t.Fatalf("expected error on rollback attempt, got nil")
	}
}
