package shell_test

import (
	"kademlia/internal/core/entities"
	"kademlia/internal/core/kademlia"
	"kademlia/internal/shell"
	"testing"
	"time"
)

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
		func(data string) (*entities.KademliaID, error) {
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

func TestNewShell(t *testing.T) {
	kademlia := newMockKademlia(t)
	shell := shell.NewShell(kademlia)
	if shell == nil {
		t.Fatalf("NewShell() returned nil")
	}
}

func TestShell_Run_NotATTY(t *testing.T) {
	kademlia := newMockKademlia(t)
	shell := shell.NewShell(kademlia)

	err := shell.Run()
	if err != nil {
		t.Fatalf("Run() returned unexpected error: %v", err)
	}
}
