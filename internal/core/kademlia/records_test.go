package kademlia

import (
	"crypto/ed25519"
	"testing"
)

func TestRecords(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("GenerateKey error: %v", err)
	}

	lp := &LatestPointer{
		Tag:               "latest-pointer",
		DomainName:        "example.com",
		PackageName:       "pkg",
		Version:           "1.0.0",
		VersionRecordHash: "hash123",
	}
	lp.Sign(priv)

	if !lp.VerifySignature(pub) {
		t.Fatalf("expected valid signature")
	}

	lp.Signature = "invalid-hex"
	if lp.VerifySignature(pub) {
		t.Fatalf("expected invalid signature on bad hex")
	}

	vr := &VersionRecord{
		Tag:                   "version-record",
		DomainName:            "example.com",
		PackageName:           "pkg",
		Version:               "1.0.0",
		BlobHash:              "blob123",
		PrevVersionRecordHash: "",
		Signature:             "not-valid-hex",
	}
	if vr.VerifySignature(pub) {
		t.Fatalf("expected false on bad hex signature")
	}

	key := GetLatestKey("example.com", "pkg")
	if key == "" {
		t.Fatalf("expected non-empty key")
	}
}
