package kademlia

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
)

var (
	ErrInvalidSignature = errors.New("invalid signature")
	ErrInvalidVersion   = errors.New("invalid version format or lower than current version")
	ErrForkDetected     = errors.New("fork detected in version history: prev hash does not match current head")
)

// VersionRecord represents an immutable package version
type VersionRecord struct {
	Tag                   string `json:"tag"` // "version-record"
	DomainName            string `json:"domain_name"`
	PackageName           string `json:"package_name"`
	Version               string `json:"version"`
	BlobHash              string `json:"blob_hash"`
	PrevVersionRecordHash string `json:"prev_version_record_hash"`
	Signature             string `json:"signature"`
}

// LatestPointer represents the mutable pointer to the latest package version
type LatestPointer struct {
	Tag               string `json:"tag"` // "latest-pointer"
	DomainName        string `json:"domain_name"`
	PackageName       string `json:"package_name"`
	Version           string `json:"version"`
	VersionRecordHash string `json:"version_record_hash"`
	Signature         string `json:"signature"`
}

// ComputePayloadHash calculates the SHA-256 hash over the payload fields for VersionRecord
func (v *VersionRecord) ComputePayloadHash() []byte {
	data := fmt.Sprintf("%s:%s:%s:%s:%s:%s",
		v.Tag, v.DomainName, v.PackageName, v.Version, v.BlobHash, v.PrevVersionRecordHash)
	hash := sha256.Sum256([]byte(data))
	return hash[:]
}

// Sign signs the VersionRecord payload with the developer's Ed25519 private key
func (v *VersionRecord) Sign(privateKey ed25519.PrivateKey) {
	payloadHash := v.ComputePayloadHash()
	sig := ed25519.Sign(privateKey, payloadHash)
	v.Signature = hex.EncodeToString(sig)
}

// VerifySignature verifies the VersionRecord signature against the developer's public key
func (v *VersionRecord) VerifySignature(publicKey ed25519.PublicKey) bool {
	sigBytes, err := hex.DecodeString(v.Signature)
	if err != nil {
		return false
	}
	payloadHash := v.ComputePayloadHash()
	return ed25519.Verify(publicKey, payloadHash, sigBytes)
}

// Hash returns the SHA-256 string identifier of the JSON-encoded VersionRecord
func (v *VersionRecord) Hash() (string, error) {
	bytes, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256(bytes)
	return hex.EncodeToString(hash[:]), nil
}

// ComputePayloadHash calculates the SHA-256 hash over the payload fields for LatestPointer
func (lp *LatestPointer) ComputePayloadHash() []byte {
	data := fmt.Sprintf("%s:%s:%s:%s:%s",
		lp.Tag, lp.DomainName, lp.PackageName, lp.Version, lp.VersionRecordHash)
	hash := sha256.Sum256([]byte(data))
	return hash[:]
}

// Sign signs the LatestPointer payload with the developer's Ed25519 private key
func (lp *LatestPointer) Sign(privateKey ed25519.PrivateKey) {
	payloadHash := lp.ComputePayloadHash()
	sig := ed25519.Sign(privateKey, payloadHash)
	lp.Signature = hex.EncodeToString(sig)
}

// VerifySignature verifies the LatestPointer signature against the developer's public key
func (lp *LatestPointer) VerifySignature(publicKey ed25519.PublicKey) bool {
	sigBytes, err := hex.DecodeString(lp.Signature)
	if err != nil {
		return false
	}
	payloadHash := lp.ComputePayloadHash()
	return ed25519.Verify(publicKey, payloadHash, sigBytes)
}

// GetLatestKey calculates the deterministic lookup key hash("domain:package:latest")
func GetLatestKey(domain, pkg string) string {
	rawKey := fmt.Sprintf("%s:%s:latest", domain, pkg)
	hash := sha256.Sum256([]byte(rawKey))
	return hex.EncodeToString(hash[:])
}
