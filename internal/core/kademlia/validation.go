package kademlia

import (
	"crypto/ed25519"
	"fmt"
	"strconv"
	"strings"
)

// ValidateNewVersion enforces business rules before accepting a package update:
// 1. Valid signature matching the domain owner's public key from DNS.
// 2. Strict version progression (prevents rollbacks).
// 3. Hash continuity with current head (prevents forks).
func ValidateNewVersion(
	newRecord *VersionRecord,
	currentHead *VersionRecord,
	pubKey ed25519.PublicKey,
) error {
	// 1. Verify cryptographic signature
	if !newRecord.VerifySignature(pubKey) {
		return ErrInvalidSignature
	}

	// First release of a package: no current head to validate against
	if currentHead == nil {
		return nil
	}

	// 2. Fork protection: the new record must reference the current head's exact hash
	currentHeadHash, err := currentHead.Hash()
	if err != nil {
		return fmt.Errorf("failed to compute current head hash: %w", err)
	}

	if newRecord.PrevVersionRecordHash != currentHeadHash {
		return ErrForkDetected
	}

	// 3. Rollback protection: version must be strictly greater than current version
	if compareVersions(newRecord.Version, currentHead.Version) <= 0 {
		return ErrInvalidVersion
	}

	return nil
}

// compareVersions compares two semver strings (e.g. "1.2.0", "v1.2.0").
// Returns:
//
//	 1 if v1 > v2
//	-1 if v1 < v2
//	 0 if v1 == v2
func compareVersions(v1, v2 string) int {
	clean := func(v string) []int {
		v = strings.TrimPrefix(v, "v")
		parts := strings.Split(v, ".")
		res := make([]int, 3)
		for i := 0; i < len(parts) && i < 3; i++ {
			num, _ := strconv.Atoi(parts[i])
			res[i] = num
		}
		return res
	}

	p1 := clean(v1)
	p2 := clean(v2)

	for i := 0; i < 3; i++ {
		if p1[i] > p2[i] {
			return 1
		}
		if p1[i] < p2[i] {
			return -1
		}
	}
	return 0
}
