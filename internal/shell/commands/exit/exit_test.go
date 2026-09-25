package exit_test

import (
	"errors"
	"kademlia/internal/core/entities"
	"kademlia/internal/core/kademlia"
	"kademlia/internal/shell/commands"
	"kademlia/internal/shell/commands/exit"
	"reflect"
	"testing"
	"time"
)

func newMockKademlia(
	t *testing.T,
	quitFunction func() error,
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
		func() []*kademlia.Bucket {
			t.Fatal("GetBucketsFunction should not be called")
			return []*kademlia.Bucket{}
		},
		func() []string {
			t.Fatal("GetStoredKeysFunction should not be called")
			return []string{}
		},
		quitFunction,
	)
}

func TestNewExitCommand(t *testing.T) {
	mockKademlia := newMockKademlia(
		t,
		func() error {
			t.Fatal("NewExitCommand() called Quit")
			return nil
		},
	)
	command := exit.NewExitCommand(mockKademlia)
	if command == nil {
		t.Fatalf("NewExitCommand() returned nil")
	}
}

func TestExitCommand_Execute_TooManyArgs(t *testing.T) {
	args := []string{"extra"}
	mockKademlia := newMockKademlia(
		t,
		func() error {
			t.Fatalf("Execute(%v) called Quit", args)
			return nil
		},
	)
	command := exit.NewExitCommand(mockKademlia)
	err := command.Execute(args)
	if err == nil {
		t.Fatalf("Execute(%v) expected error, got nil", args)
	}
}

func TestExitCommand_Execute_CallsQuit(t *testing.T) {
	quitCalled := false
	args := []string{}
	mockKademlia := newMockKademlia(
		t,
		func() error {
			quitCalled = true
			return nil
		},
	)
	command := exit.NewExitCommand(mockKademlia)
	err := command.Execute(args)
	if err != nil {
		t.Fatalf("Execute(%v) returned unexpected error: %v", args, err)
	}
	if !quitCalled {
		t.Fatalf("Execute(%v) didn't call Quit", args)
	}
}

func TestExitCommand_Execute_QuitFail(t *testing.T) {
	quitCalled := false
	args := []string{}
	expectedError := errors.New("test error")
	mockKademlia := newMockKademlia(
		t,
		func() error {
			quitCalled = true
			return expectedError
		},
	)
	command := exit.NewExitCommand(mockKademlia)
	err := command.Execute(args)
	if !errors.Is(err, expectedError) {
		t.Fatalf("Execute(%v) err = %v; want %v", args, err, expectedError)
	}
	if !quitCalled {
		t.Fatalf("Execute(%v) didn't call Quit", args)
	}
}

func TestExitCommand_GetCompletions(t *testing.T) {
	mockKademlia := newMockKademlia(
		t,
		func() error {
			t.Fatal("GetCompletions() called Quit")
			return nil
		},
	)
	command := exit.NewExitCommand(mockKademlia)
	completions := command.GetCompletions()
	expectedCompletions := []commands.Completion{
		{Name: "--help"},
		{Name: "-h"},
	}
	if completions == nil || !reflect.DeepEqual(
		completions,
		expectedCompletions,
	) {
		t.Fatalf(
			"GetCompletions() completions = %v; want %v",
			completions,
			expectedCompletions,
		)
	}
}
