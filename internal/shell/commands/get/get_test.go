package get_test

import (
	"bytes"
	"kademlia/internal/core/entities"
	"kademlia/internal/core/kademlia"
	"kademlia/internal/shell/commands/get"
	"testing"
	"time"
)

func newMockKademlia(
	t *testing.T,
	getValueFunction func(string) (*string, *string, error),
) kademlia.Kademlia {
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
		func() []string {
			t.Fatal("GetStoredKeysFunction should not be called")
			return []string{}
		},
		getValueFunction,
		func() error {
			t.Fatal("Quit should not be called")
			return nil
		},
	)
}

func TestNewGetCommand(t *testing.T) {
	mockKademlia := newMockKademlia(
		t,
		func(string) (*string, *string, error) {
			t.Fatal("NewGetCommand() called GetValue")
			return nil, nil, nil
		},
	)
	var mockOut bytes.Buffer
	command := get.NewGetCommand(mockKademlia, &mockOut)
	if command == nil {
		t.Fatalf("NewGetCommand() returned nil")
	}
}

func TestGetCommand_Execute_TooManyArgs(t *testing.T) {
	args := []string{"key", "filename", "extra"}
	mockKademlia := newMockKademlia(
		t,
		func(string) (*string, *string, error) {
			t.Fatal("NewGetCommand() called GetValue")
			return nil, nil, nil
		},
	)
	var mockOut bytes.Buffer
	command := get.NewGetCommand(mockKademlia, &mockOut)
	err := command.Execute(args)
	if err == nil {
		t.Fatalf("Execute(%v) expected error, got nil", args)
	}
}
