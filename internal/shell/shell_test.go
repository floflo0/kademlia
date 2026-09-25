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
		func() []*kademlia.Bucket {
			t.Fatal("GetBucketsFunction should not be called")
			return []*kademlia.Bucket{}
		},
		func() []string {
			t.Fatal("GetStoredKeysFunction should not be called")
			return []string{}
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
