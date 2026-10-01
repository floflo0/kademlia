package put_test

import (
	"bytes"
	"errors"
	"kademlia/internal/core/entities"
	"kademlia/internal/core/kademlia"
	"kademlia/internal/shell/commands"
	"kademlia/internal/shell/commands/put"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func newMockKademlia(
	t *testing.T,
	putFunction func(data string) (*entities.KademliaID, error),
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
		putFunction,
		func() []*kademlia.Bucket {
			t.Fatal("GetBucketsFunction should not be called")
			return []*kademlia.Bucket{}
		},
		func() []string {
			t.Fatal("GetStoredKeysFunction should not be called")
			return []string{}
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

func createTestFile(t *testing.T) string {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "file.txt")
	if err := os.WriteFile(filePath, []byte("hello"), 0o644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	return filePath
}

func TestNewPutCommand(t *testing.T) {
	mockKademlia := newMockKademlia(
		t,
		func(value string) (*entities.KademliaID, error) {
			t.Fatal("NewPingCommand() called Put")
			return nil, nil
		},
	)
	var mockOut bytes.Buffer
	command := put.NewPutCommand(mockKademlia, &mockOut)
	if command == nil {
		t.Fatalf("NewPutCommand() returned nil")
	}
}

func TestPutCommand_Execute_NoArgs(t *testing.T) {
	args := []string{}
	kademlia := newMockKademlia(
		t,
		func(value string) (*entities.KademliaID, error) {
			t.Fatalf("Execute(%v) called Put", args)
			return nil, nil
		},
	)
	var mockOut bytes.Buffer
	command := put.NewPutCommand(kademlia, &mockOut)
	err := command.Execute(args)
	if err == nil {
		t.Fatalf("Execute(%v) expected error, got nil", args)
	}
}

func TestPutCommand_Execute_TooManyArgs(t *testing.T) {
	args := []string{"file.txt", "extra"}
	mockKademlia := newMockKademlia(
		t,
		func(value string) (*entities.KademliaID, error) {
			t.Fatalf("Execute(%v) called Put", args)
			return nil, nil
		},
	)
	var mockOut bytes.Buffer
	command := put.NewPutCommand(mockKademlia, &mockOut)
	err := command.Execute(args)
	expectedError := "accepts 1 arg(s), received 2"
	if err == nil || err.Error() != expectedError {
		t.Fatalf("Execute(%v) err = %v; want %v", args, err, expectedError)
	}
}

func TestPutCommand_Execute_ReadFile_Error(t *testing.T) {
	args := []string{"file.txt"}
	mockKademlia := newMockKademlia(
		t,
		func(value string) (*entities.KademliaID, error) {
			t.Fatalf("Execute(%v) called Put", args)
			return nil, nil
		},
	)
	var mockOut bytes.Buffer
	command := put.NewPutCommand(mockKademlia, &mockOut)
	err := command.Execute(args)
	expectedError := os.ErrNotExist
	if !errors.Is(err, expectedError) {
		t.Fatalf("Execute(%v) err = %v; want %v", args, err, expectedError)
	}
}

func TestPutCommand_Execute_Put_Sucess(t *testing.T) {
	filePath := createTestFile(t)
	args := []string{filePath}
	putCalled := false
	id, err := entities.NewKademliaID(
		"0000000000000000000000000000000000000000000000000000000000000000",
	)
	if err != nil {
		t.Fatalf("NewKademliaID() returns unexpected error: %v", err)
	}
	mockKademlia := newMockKademlia(
		t,
		func(value string) (*entities.KademliaID, error) {
			putCalled = true
			return id, nil
		},
	)
	var mockOut bytes.Buffer
	command := put.NewPutCommand(mockKademlia, &mockOut)
	err = command.Execute(args)
	if err != nil {
		t.Fatalf("Execute(%v) returned unexpected error: %v", args, err)
	}
	if !putCalled {
		t.Fatalf("Execute(%v) didn't call Put", args)
	}
	expectedOutput := "File successfully stored with ID = " +
		"\"0000000000000000000000000000000000000000000000000000000000000000\"\n"
	output := mockOut.String()
	if output != expectedOutput {
		t.Fatalf(
			"Execute(%v) expected output = %q but got %q",
			args,
			expectedOutput,
			output,
		)
	}
}

func TestPutCommand_Execute_Put_Error(t *testing.T) {
	filePath := createTestFile(t)
	args := []string{filePath}
	putCalled := false
	expectedError := errors.New("test errror")
	mockKademlia := newMockKademlia(
		t,
		func(value string) (*entities.KademliaID, error) {
			putCalled = true
			return nil, expectedError
		},
	)
	var mockOut bytes.Buffer
	command := put.NewPutCommand(mockKademlia, &mockOut)
	err := command.Execute(args)
	if !errors.Is(err, expectedError) {
		t.Fatalf("Execute(%v) err = %v; want %v", args, err, expectedError)
	}
	if !putCalled {
		t.Fatalf("Execute(%v) didn't call Put", args)
	}
}

func TestPutCommand_GetCompletions(t *testing.T) {
	mockKademlia := newMockKademlia(
		t,
		func(value string) (*entities.KademliaID, error) {
			t.Fatal("GetCompletions() called Put")
			return nil, nil
		},
	)
	var mockOut bytes.Buffer
	command := put.NewPutCommand(mockKademlia, &mockOut)
	completions := command.GetCompletions()
	exepectedCompletions := []commands.Completion{
		{Name: "--help"},
		{Name: "-h"},
	}
	if completions == nil || !reflect.DeepEqual(
		completions,
		exepectedCompletions,
	) {
		t.Fatalf(
			"GetCompletions() completions = %v; want %v",
			completions,
			exepectedCompletions,
		)
	}
}
