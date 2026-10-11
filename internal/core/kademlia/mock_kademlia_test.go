package kademlia_test

import (
	"crypto/ed25519"
	"kademlia/internal/core/entities"
	"kademlia/internal/core/kademlia"
	"kademlia/internal/core/ports"
	"testing"
	"time"
)

func TestNewMockKademlia(t *testing.T) {
	kademlia := kademlia.NewMockKademlia(
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
		func(
			domain string,
			packageName string,
			version string,
			blob string,
			privKey ed25519.PrivateKey,
			dnsVerifier ports.DNSVerifier,
			force bool,
			explicitPrev string,
		) (*kademlia.VersionRecord, error) {
			t.Fatal("PublishPackage should not be called")
			return nil, nil
		},
		func(
			domain string,
			packageName string,
			version string,
		) (string, string, error) {
			t.Fatal("InstallPackage should not be called")
			return "", "", nil
		},
		func(
			domain string,
			packageName string,
		) ([]*kademlia.VersionRecord, error) {
			t.Fatal("GetVersionChain should not be called")
			return nil, nil
		},
	)
	if kademlia == nil {
		t.Fatalf("NewMockKademlia() returns nil")
	}
}

func TestNewMockKademlia_Run(t *testing.T) {
	runCalled := false
	kademlia := kademlia.NewMockKademlia(
		func(firstContact *entities.Address) error {
			runCalled = true
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
		func(
			domain string,
			packageName string,
			version string,
			blob string,
			privKey ed25519.PrivateKey,
			dnsVerifier ports.DNSVerifier,
			force bool,
			explicitPrev string,
		) (*kademlia.VersionRecord, error) {
			t.Fatal("PublishPackage should not be called")
			return nil, nil
		},
		func(
			domain string,
			packageName string,
			version string,
		) (string, string, error) {
			t.Fatal("InstallPackage should not be called")
			return "", "", nil
		},
		func(
			domain string,
			packageName string,
		) ([]*kademlia.VersionRecord, error) {
			t.Fatal("GetVersionChain should not be called")
			return nil, nil
		},
	)
	kademlia.Run(nil)
	if !runCalled {
		t.Fatalf("Run() didn't call the run function")
	}
}

func TestNewMockKademlia_Ping(t *testing.T) {
	pingCalled := false
	kademlia := kademlia.NewMockKademlia(
		func(firstContact *entities.Address) error {
			t.Fatal("Run should not be called")
			return nil
		},
		func(address entities.Address) (time.Duration, error) {
			pingCalled = true
			return 0, nil
		},
		func(value string) (*entities.KademliaID, error) {
			t.Fatal("Put should not be called")
			return nil, nil
		},
		func() []*kademlia.Bucket {
			t.Fatal("GetBuckets should not be called")
			return nil
		},
		func() []entities.KademliaID {
			t.Fatal("GetStoredKeys should not be called")
			return nil
		},
		func(string) (*string, *string, error) {
			t.Fatal("GetValue should not be called")
			return nil, nil, nil
		},
		func() error {
			t.Fatal("Quit should not be called")
			return nil
		},
		func(
			domain string,
			packageName string,
			version string,
			blob string,
			privKey ed25519.PrivateKey,
			dnsVerifier ports.DNSVerifier,
			force bool,
			explicitPrev string,
		) (*kademlia.VersionRecord, error) {
			t.Fatal("PublishPackage should not be called")
			return nil, nil
		},
		func(
			domain string,
			packageName string,
			version string,
		) (string, string, error) {
			t.Fatal("InstallPackage should not be called")
			return "", "", nil
		},
		func(
			domain string,
			packageName string,
		) ([]*kademlia.VersionRecord, error) {
			t.Fatal("GetVersionChain should not be called")
			return nil, nil
		},
	)
	kademlia.Ping(entities.Address{})
	if !pingCalled {
		t.Fatalf("Ping() didn't call the ping function")
	}
}

func TestNewMockKademlia_Put(t *testing.T) {
	putCalled := false
	kademila := kademlia.NewMockKademlia(
		func(firstContact *entities.Address) error {
			t.Fatal("Run should not be called")
			return nil
		},
		func(address entities.Address) (time.Duration, error) {
			t.Fatal("Ping should not be called")
			return 0, nil
		},
		func(value string) (*entities.KademliaID, error) {
			putCalled = true
			return nil, nil
		},
		func() []*kademlia.Bucket {
			t.Fatal("GetBuckets should not be called")
			return nil
		},
		func() []entities.KademliaID {
			t.Fatal("GetStoredKeys should not be called")
			return nil
		},
		func(string) (*string, *string, error) {
			t.Fatal("GetValue should not be called")
			return nil, nil, nil
		},
		func() error {
			t.Fatal("Quit should not be called")
			return nil
		},
		func(
			domain string,
			packageName string,
			version string,
			blob string,
			privKey ed25519.PrivateKey,
			dnsVerifier ports.DNSVerifier,
			force bool,
			explicitPrev string,
		) (*kademlia.VersionRecord, error) {
			t.Fatal("PublishPackage should not be called")
			return nil, nil
		},
		func(
			domain string,
			packageName string,
			version string,
		) (string, string, error) {
			t.Fatal("InstallPackage should not be called")
			return "", "", nil
		},
		func(
			domain string,
			packageName string,
		) ([]*kademlia.VersionRecord, error) {
			t.Fatal("GetVersionChain should not be called")
			return nil, nil
		},
	)
	kademila.Put("")
	if !putCalled {
		t.Fatalf("Put() didn't call the put function")
	}
}

func TestNewMockKademlia_GetBuckets(t *testing.T) {
	getBucketsCalled := false
	kademlia := kademlia.NewMockKademlia(
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
			getBucketsCalled = true
			return nil
		},
		func() []entities.KademliaID {
			t.Fatal("GetStoredKeys should not be called")
			return nil
		},
		func(string) (*string, *string, error) {
			t.Fatal("GetValue should not be called")
			return nil, nil, nil
		},
		func() error {
			t.Fatal("Quit should not be called")
			return nil
		},
		func(
			domain string,
			packageName string,
			version string,
			blob string,
			privKey ed25519.PrivateKey,
			dnsVerifier ports.DNSVerifier,
			force bool,
			explicitPrev string,
		) (*kademlia.VersionRecord, error) {
			t.Fatal("PublishPackage should not be called")
			return nil, nil
		},
		func(
			domain string,
			packageName string,
			version string,
		) (string, string, error) {
			t.Fatal("InstallPackage should not be called")
			return "", "", nil
		},
		func(
			domain string,
			packageName string,
		) ([]*kademlia.VersionRecord, error) {
			t.Fatal("GetVersionChain should not be called")
			return nil, nil
		},
	)
	kademlia.GetBuckets()
	if !getBucketsCalled {
		t.Fatalf("GetBuckets() didn't call the getBuckets function")
	}
}

func TestNewMockKademlia_GetStoredKeys(t *testing.T) {
	getStoredKeysCalled := false
	kademlia := kademlia.NewMockKademlia(
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
			t.Fatal("GetBuckets should not be called")
			return nil
		},
		func() []entities.KademliaID {
			getStoredKeysCalled = true
			return nil
		},
		func(string) (*string, *string, error) {
			t.Fatal("GetValue should not be called")
			return nil, nil, nil
		},
		func() error {
			t.Fatal("Quit should not be called")
			return nil
		},
		func(
			domain string,
			packageName string,
			version string,
			blob string,
			privKey ed25519.PrivateKey,
			dnsVerifier ports.DNSVerifier,
			force bool,
			explicitPrev string,
		) (*kademlia.VersionRecord, error) {
			t.Fatal("PublishPackage should not be called")
			return nil, nil
		},
		func(
			domain string,
			packageName string,
			version string,
		) (string, string, error) {
			t.Fatal("InstallPackage should not be called")
			return "", "", nil
		},
		func(
			domain string,
			packageName string,
		) ([]*kademlia.VersionRecord, error) {
			t.Fatal("GetVersionChain should not be called")
			return nil, nil
		},
	)
	kademlia.GetStoredKeys()
	if !getStoredKeysCalled {
		t.Fatalf("GetStoredKeys() didn't call the getStoredKeys function")
	}
}

func TestNewMockKademlia_GetValue(t *testing.T) {
	getValueCalled := false
	kademlia := kademlia.NewMockKademlia(
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
			t.Fatal("GetBuckets should not be called")
			return nil
		},
		func() []entities.KademliaID {
			t.Fatal("GetStoredKeys should not be called")
			return nil
		},
		func(string) (*string, *string, error) {
			getValueCalled = true
			return nil, nil, nil
		},
		func() error {
			t.Fatal("Quit should not be called")
			return nil
		},
		func(
			domain string,
			packageName string,
			version string,
			blob string,
			privKey ed25519.PrivateKey,
			dnsVerifier ports.DNSVerifier,
			force bool,
			explicitPrev string,
		) (*kademlia.VersionRecord, error) {
			t.Fatal("PublishPackage should not be called")
			return nil, nil
		},
		func(
			domain string,
			packageName string,
			version string,
		) (string, string, error) {
			t.Fatal("InstallPackage should not be called")
			return "", "", nil
		},
		func(
			domain string,
			packageName string,
		) ([]*kademlia.VersionRecord, error) {
			t.Fatal("GetVersionChain should not be called")
			return nil, nil
		},
	)
	kademlia.GetValue("")
	if !getValueCalled {
		t.Fatalf("GetValue() didn't call the getValue function")
	}
}

func TestNewMockKademlia_Quit(t *testing.T) {
	quitCalled := false
	kademlia := kademlia.NewMockKademlia(
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
			t.Fatal("GetBuckets should not be called")
			return nil
		},
		func() []entities.KademliaID {
			t.Fatal("GetStoredKeys should not be called")
			return nil
		},
		func(string) (*string, *string, error) {
			t.Fatal("GetValue should not be called")
			return nil, nil, nil
		},
		func() error {
			quitCalled = true
			return nil
		},
		func(
			domain string,
			packageName string,
			version string,
			blob string,
			privKey ed25519.PrivateKey,
			dnsVerifier ports.DNSVerifier,
			force bool,
			explicitPrev string,
		) (*kademlia.VersionRecord, error) {
			t.Fatal("PublishPackage should not be called")
			return nil, nil
		},
		func(
			domain string,
			packageName string,
			version string,
		) (string, string, error) {
			t.Fatal("InstallPackage should not be called")
			return "", "", nil
		},
		func(
			domain string,
			packageName string,
		) ([]*kademlia.VersionRecord, error) {
			t.Fatal("GetVersionChain should not be called")
			return nil, nil
		},
	)
	kademlia.Quit()
	if !quitCalled {
		t.Fatalf("Quit() didn't call the quit function")
	}
}

func TestNewMockKademlia_PublishPackage(t *testing.T) {
	publishPackageCalled := false
	kademlia := kademlia.NewMockKademlia(
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
			t.Fatal("GetBuckets should not be called")
			return nil
		},
		func() []entities.KademliaID {
			t.Fatal("GetStoredKeys should not be called")
			return nil
		},
		func(string) (*string, *string, error) {
			t.Fatal("GetValue should not be called")
			return nil, nil, nil
		},
		func() error {
			t.Fatal("Quit should not be called")
			return nil
		},
		func(
			domain string,
			packageName string,
			version string,
			blob string,
			privKey ed25519.PrivateKey,
			dnsVerifier ports.DNSVerifier,
			force bool,
			explicitPrev string,
		) (*kademlia.VersionRecord, error) {
			publishPackageCalled = true
			return nil, nil
		},
		func(
			domain string,
			packageName string,
			version string,
		) (string, string, error) {
			t.Fatal("InstallPackage should not be called")
			return "", "", nil
		},
		func(
			domain string,
			packageName string,
		) ([]*kademlia.VersionRecord, error) {
			t.Fatal("GetVersionChain should not be called")
			return nil, nil
		},
	)
	kademlia.PublishPackage("", "", "", "", nil, nil, false, "")
	if !publishPackageCalled {
		t.Fatalf("PublishPackage() didn't call the publishPackage function")
	}
}

func TestNewMockKademlia_InstallPackage(t *testing.T) {
	installPackageCalled := false
	kademlia := kademlia.NewMockKademlia(
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
			t.Fatal("GetBuckets should not be called")
			return nil
		},
		func() []entities.KademliaID {
			t.Fatal("GetStoredKeys should not be called")
			return nil
		},
		func(string) (*string, *string, error) {
			t.Fatal("GetValue should not be called")
			return nil, nil, nil
		},
		func() error {
			t.Fatal("Quit should not be called")
			return nil
		},
		func(
			domain string,
			packageName string,
			version string,
			blob string,
			privKey ed25519.PrivateKey,
			dnsVerifier ports.DNSVerifier,
			force bool,
			explicitPrev string,
		) (*kademlia.VersionRecord, error) {
			t.Fatal("PublishPackage should not be called")
			return nil, nil
		},
		func(
			domain string,
			packageName string,
			version string,
		) (string, string, error) {
			installPackageCalled = true
			return "", "", nil
		},
		func(
			domain string,
			packageName string,
		) ([]*kademlia.VersionRecord, error) {
			t.Fatal("GetVersionChain should not be called")
			return nil, nil
		},
	)
	kademlia.InstallPackage("", "", "")
	if !installPackageCalled {
		t.Fatalf("InstallPackage() didn't call the installPackage function")
	}
}

func TestNewMockKademlia_GetVersionChain(t *testing.T) {
	getVersionChainCalled := false
	kademlia := kademlia.NewMockKademlia(
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
			t.Fatal("GetBuckets should not be called")
			return nil
		},
		func() []entities.KademliaID {
			t.Fatal("GetStoredKeys should not be called")
			return nil
		},
		func(string) (*string, *string, error) {
			t.Fatal("GetValue should not be called")
			return nil, nil, nil
		},
		func() error {
			t.Fatal("Quit should not be called")
			return nil
		},
		func(
			domain string,
			packageName string,
			version string,
			blob string,
			privKey ed25519.PrivateKey,
			dnsVerifier ports.DNSVerifier,
			force bool,
			explicitPrev string,
		) (*kademlia.VersionRecord, error) {
			t.Fatal("PublishPackage should not be called")
			return nil, nil
		},
		func(
			domain string,
			packageName string,
			version string,
		) (string, string, error) {
			t.Fatal("InstallPackage should not be called")
			return "", "", nil
		},
		func(
			domain string,
			packageName string,
		) ([]*kademlia.VersionRecord, error) {
			getVersionChainCalled = true
			return nil, nil
		},
	)
	kademlia.GetVersionChain("", "")
	if !getVersionChainCalled {
		t.Fatalf("GetVersionChain() didn't call the getVersionChain function")
	}
}
