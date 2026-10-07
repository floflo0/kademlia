package kademlia

import (
	"crypto/ed25519"
	"kademlia/internal/core/entities"
	"time"
)

type mockKademlia struct {
	runFunction           func(firstContact *entities.Address) error
	putFunction           func(data string) (*entities.KademliaID, error)
	pingFunction          func(address entities.Address) (time.Duration, error)
	getBucketsFunction    func() []*Bucket
	getStoredKeysFunction func() []entities.KademliaID
	getValueFunction      func(key string) (*string, *string, error)
	quitFunction          func() error
}

func NewMockKademlia(
	runFunction func(firstContact *entities.Address) error,
	pingFunction func(address entities.Address) (time.Duration, error),
	putFunction func(data string) (*entities.KademliaID, error),
	getBucketsFunction func() []*Bucket,
	getStoredKeysFunction func() []entities.KademliaID,
	getValueFunction func(key string) (*string, *string, error),
	quitFunction func() error,
) Kademlia {
	return &mockKademlia{
		runFunction:           runFunction,
		pingFunction:          pingFunction,
		putFunction:           putFunction,
		getBucketsFunction:    getBucketsFunction,
		getStoredKeysFunction: getStoredKeysFunction,
		getValueFunction:      getValueFunction,
		quitFunction:          quitFunction,
	}
}

func (k *mockKademlia) Run(firstContact *entities.Address) error {
	return k.runFunction(firstContact)
}

func (k *mockKademlia) Ping(address entities.Address) (time.Duration, error) {
	return k.pingFunction(address)
}

func (k *mockKademlia) Put(data string) (*entities.KademliaID, error) {
	return k.putFunction(data)
}

func (k *mockKademlia) GetBuckets() []*Bucket {
	return k.getBucketsFunction()
}

func (k *mockKademlia) GetStoredKeys() []entities.KademliaID {
	return k.getStoredKeysFunction()
}

func (k *mockKademlia) GetValue(key string) (*string, *string, error) {
	return k.getValueFunction(key)
}

func (k *mockKademlia) Quit() error {
	return k.quitFunction()
}

func (m *mockKademlia) PublishPackage(
	domain string,
	packageName string,
	version string,
	blob string,
	privKey ed25519.PrivateKey,
	dns DNSVerifier,
	force bool,
	explicitPrev string,
) (*VersionRecord, error) {
	return &VersionRecord{
		DomainName:  domain,
		PackageName: packageName,
		Version:     version,
		BlobHash:    blob,
	}, nil
}

func (m *mockKademlia) InstallPackage(domain, packageName, version string) (string, string, error) {
	return "mock-blob-content", version, nil
}

func (m *mockKademlia) GetVersionChain(domain, packageName string) ([]*VersionRecord, error) {
	return []*VersionRecord{}, nil
}
