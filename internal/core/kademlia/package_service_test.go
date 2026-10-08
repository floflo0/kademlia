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
	_, err = k.PublishPackage(domain, pkg, "1.0.0", blobV1, privKey, dns, false, "")
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
	_, err = k.PublishPackage(domain, pkg, "1.1.0", blobV2, privKey, dns, false, "")
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
	_, err = k.PublishPackage(domain, pkg, "1.0.5", "invalid_rollback", privKey, dns, false, "")
	if err == nil {
		t.Fatalf("expected error on rollback attempt, got nil")
	}
}

func TestPublishPackage_WithForceAndPrev(t *testing.T) {
	k := newTestKademliaNode()

	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate keys: %v", err)
	}

	dns := NewMockDNSVerifier()
	dns.RegisterDomain("example.com", pubKey)

	domain := "example.com"
	pkg := "my-lib"

	// 1. Publish version 1.0.0 (normal)
	_, err = k.PublishPackage(domain, pkg, "1.0.0", "content1", privKey, dns, false, "")
	if err != nil {
		t.Fatalf("Error publishing v1.0.0: %v", err)
	}

	// 2. Try to publish a smaller version (0.9.0) without force -> Should fail
	_, err = k.PublishPackage(domain, pkg, "0.9.0", "content0", privKey, dns, false, "")
	if err == nil {
		t.Errorf("Fail was expected when publishing minor version without --force, but was an exit")
	}

	// 3. Publicar versión menor (0.9.0) CON force -> Debe tener éxito
	_, err = k.PublishPackage(domain, pkg, "0.9.0", "content0", privKey, dns, true, "")
	if err != nil {
		t.Errorf("Publishing failed with --force: %v", err)
	}

	// 4. Probar GetVersionChain
	chain, err := k.GetVersionChain(domain, pkg)
	if err != nil {
		t.Fatalf("Error obtained version's chain: %v", err)
	}
	if len(chain) == 0 {
		t.Errorf(" The version's chain shouldn't be empty")
	}
}
