package kademlia

import (
	"crypto/ed25519"
	"encoding/json"
	"errors"
	"fmt"
	"kademlia/internal/core/entities"
)

var (
	ErrPackageNotFound = errors.New("package or requested version not found in history")
)

// storeData stores the value in the local CAS dataStore and safely invokes network Put if RoutingTable exists.
func (k *kademlia) storeData(data string) (entities.KademliaID, error) {
	idPtr := entities.NewKademliaIDFromString(data)
	var id entities.KademliaID
	if idPtr != nil {
		id = *idPtr
	}

	if k.dataStore != nil {
		if err := k.dataStore.Put(id, data); err != nil {
			return id, err
		}
	}
	if k.RoutingTable != nil {
		if _, err := k.Put(data); err != nil {
			return id, err
		}
	}
	return id, nil
}

// PublishPackage handles signing, validating, storing the blob, storing the VersionRecord,
// and updating the LatestPointer in the Kademlia network.
// Soporta los flags --force (force) y --prev (explicitPrev).
func (k *kademlia) PublishPackage(
	domain string,
	packageName string,
	version string,
	blob string,
	privKey ed25519.PrivateKey,
	dns DNSVerifier,
	force bool,
	explicitPrev string,
) (*VersionRecord, error) {
	// 1. Verify domain ownership via DNS
	pubKey, err := dns.GetPublicKey(domain)
	if err != nil {
		return nil, fmt.Errorf("failed to verify domain ownership via DNS: %w", err)
	}

	// 2. Store Blob content in DHT (key = Hash(blob))
	blobID, err := k.storeData(blob)
	if err != nil {
		return nil, fmt.Errorf("failed to store package blob: %w", err)
	}
	blobHash := blobID.String()

	// 3. Determine PrevVersionRecordHash and currentHead
	var currentHead *VersionRecord
	var prevRecordHash string

	if explicitPrev != "" {
		// Si el usuario especificó --prev=VERSION
		prevHead, err := k.findVersionRecordByVersion(domain, packageName, explicitPrev)
		if err == nil && prevHead != nil {
			currentHead = prevHead
			prevHash, _ := prevHead.Hash()
			prevRecordHash = prevHash
		} else if !force {
			return nil, fmt.Errorf("specified previous version %s not found", explicitPrev)
		}
	} else {
		// Comportamiento por defecto: usar la versión 'latest' actual como anterior
		latestPtr, err := k.findLatestPointer(domain, packageName)
		if err == nil && latestPtr != nil {
			prevRecordHash = latestPtr.VersionRecordHash
			if headRec, err := k.findVersionRecordByHash(prevRecordHash); err == nil {
				currentHead = headRec
			}
		}
	}

	// 4. Construct new VersionRecord
	newRecord := &VersionRecord{
		Tag:                   "version-record",
		DomainName:            domain,
		PackageName:           packageName,
		Version:               version,
		BlobHash:              blobHash,
		PrevVersionRecordHash: prevRecordHash,
	}
	newRecord.Sign(privKey)

	// 5. Enforce business validation rules (OMITIR SI force == true)
	if !force {
		if err := ValidateNewVersion(newRecord, currentHead, pubKey); err != nil {
			return nil, fmt.Errorf("package validation failed: %w", err)
		}
	}

	// 6. Store VersionRecord in DHT
	recordJSON, err := json.Marshal(newRecord)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal version record: %w", err)
	}

	recID, err := k.storeData(string(recordJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to store version record: %w", err)
	}
	recordHash := recID.String()

	// 7. Create and store LatestPointer in DHT
	newPointer := &LatestPointer{
		Tag:               "latest-pointer",
		DomainName:        domain,
		PackageName:       packageName,
		Version:           version,
		VersionRecordHash: recordHash,
	}
	newPointer.Sign(privKey)

	pointerJSON, err := json.Marshal(newPointer)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal latest pointer: %w", err)
	}

	if _, err := k.storeData(string(pointerJSON)); err != nil {
		return nil, fmt.Errorf("failed to store latest pointer: %w", err)
	}

	return newRecord, nil
}

// InstallPackage resolves and fetches the blob content for a package.
func (k *kademlia) InstallPackage(domain, packageName, version string) (string, string, error) {
	latestPtr, err := k.findLatestPointer(domain, packageName)
	if err != nil || latestPtr == nil {
		return "", "", fmt.Errorf("%w: pointer missing for %s:%s", ErrPackageNotFound, domain, packageName)
	}

	currentHash := latestPtr.VersionRecordHash

	// Walk backwards through version history via PrevVersionRecordHash
	for currentHash != "" {
		rec, err := k.findVersionRecordByHash(currentHash)
		if err != nil || rec == nil {
			return "", "", fmt.Errorf("%w: version record missing for hash %s", ErrPackageNotFound, currentHash)
		}

		if version == "" || version == "latest" || rec.Version == version {
			blob, err := k.findBlobByHash(rec.BlobHash)
			if err != nil || blob == "" {
				return "", "", fmt.Errorf("%w: blob missing for hash %s", ErrPackageNotFound, rec.BlobHash)
			}
			return blob, rec.Version, nil
		}

		currentHash = rec.PrevVersionRecordHash
	}

	return "", "", fmt.Errorf("%w: version %s not found for %s:%s", ErrPackageNotFound, version, domain, packageName)
}

// Helper methods to query the CAS store directly
func (k *kademlia) findLatestPointer(domain, packageName string) (*LatestPointer, error) {
	if k.dataStore == nil {
		return nil, ErrPackageNotFound
	}
	var latest *LatestPointer
	for _, key := range k.dataStore.Keys() {
		val, err := k.dataStore.Get(key)
		if err != nil {
			continue
		}
		var ptr LatestPointer
		if err := json.Unmarshal([]byte(val), &ptr); err == nil && ptr.Tag == "latest-pointer" {
			if ptr.DomainName == domain && ptr.PackageName == packageName {
				if latest == nil || compareVersions(ptr.Version, latest.Version) > 0 {
					latest = &ptr
				}
			}
		}
	}
	if latest == nil {
		return nil, ErrPackageNotFound
	}
	return latest, nil
}

func (k *kademlia) findVersionRecordByHash(hashStr string) (*VersionRecord, error) {
	if k.dataStore == nil {
		return nil, ErrPackageNotFound
	}
	for _, key := range k.dataStore.Keys() {
		val, err := k.dataStore.Get(key)
		if err != nil {
			continue
		}
		var rec VersionRecord
		if err := json.Unmarshal([]byte(val), &rec); err == nil && rec.Tag == "version-record" {
			computedHash, _ := rec.Hash()
			if key.String() == hashStr || computedHash == hashStr {
				return &rec, nil
			}
		}
	}
	return nil, ErrPackageNotFound
}

func (k *kademlia) findBlobByHash(hashStr string) (string, error) {
	if k.dataStore == nil {
		return "", ErrPackageNotFound
	}
	for _, key := range k.dataStore.Keys() {
		if key.String() == hashStr {
			val, err := k.dataStore.Get(key)
			if err == nil {
				return val, nil
			}
		}
	}
	return "", ErrPackageNotFound
}

// GetVersionChain recovers all the version's history in reverse orden
func (k *kademlia) GetVersionChain(domain, packageName string) ([]*VersionRecord, error) {
	latestPtr, err := k.findLatestPointer(domain, packageName)
	if err != nil || latestPtr == nil {
		return nil, fmt.Errorf("%w: package %s:%s not found", ErrPackageNotFound, domain, packageName)
	}

	var chain []*VersionRecord
	currentHash := latestPtr.VersionRecordHash
	visited := make(map[string]bool) // to prevent infinite loops

	for currentHash != "" && !visited[currentHash] {
		visited[currentHash] = true
		rec, err := k.findVersionRecordByHash(currentHash)
		if err != nil || rec == nil {
			break
		}
		chain = append(chain, rec)
		currentHash = rec.PrevVersionRecordHash
	}

	return chain, nil
}

// Aux helper to find a VersionRecord by its specific version number
func (k *kademlia) findVersionRecordByVersion(domain, packageName, version string) (*VersionRecord, error) {
	if k.dataStore == nil {
		return nil, ErrPackageNotFound
	}
	for _, key := range k.dataStore.Keys() {
		val, err := k.dataStore.Get(key)
		if err != nil {
			continue
		}
		var rec VersionRecord
		if err := json.Unmarshal([]byte(val), &rec); err == nil && rec.Tag == "version-record" {
			if rec.DomainName == domain && rec.PackageName == packageName && rec.Version == version {
				return &rec, nil
			}
		}
	}
	return nil, ErrPackageNotFound
}
