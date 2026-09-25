package show_test

import (
	"bytes"
	"kademlia/internal/core/entities"
	"kademlia/internal/core/kademlia"
	"kademlia/internal/shell/commands"
	"kademlia/internal/shell/commands/show"
	"reflect"
	"strings"
	"testing"
	"time"
)

func newMockKademlia(
	t *testing.T,
	getBucketsFunction func() []*kademlia.Bucket,
	getStoredKeysFunction func() []string,
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
		getBucketsFunction,
		getStoredKeysFunction,
		func() error {
			t.Fatal("Quit should not be called")
			return nil
		},
	)
}

func TestNewShowCommand(t *testing.T) {
	mockKademlia := newMockKademlia(
		t,
		func() []*kademlia.Bucket {
			t.Fatal("NewShowCommand() called GetBuckets")
			return nil
		},
		func() []string {
			t.Fatal("NewShowCommand() called GetStoredKeys")
			return nil
		},
	)
	var mockOut bytes.Buffer
	command := show.NewShowCommand(mockKademlia, &mockOut)
	if command == nil {
		t.Fatalf("NewShowCommand() returned nil")
	}
}

func TestShowCommand_Execute_NoArguments(t *testing.T) {
	args := []string{}
	mockKademlia := newMockKademlia(
		t,
		func() []*kademlia.Bucket {
			t.Fatalf("Execute(%v) called GetBuckets", args)
			return nil
		},
		func() []string {
			t.Fatalf("Execute(%v) called GetStoredKeys", args)
			return nil
		},
	)
	var mockOut bytes.Buffer
	command := show.NewShowCommand(mockKademlia, &mockOut)
	err := command.Execute(args)
	if err == nil || !strings.Contains(err.Error(), "missing subcommand") {
		t.Fatalf("Execute(%v) err = %v; want missing subcommand", args, err)
	}
}

func TestShowCommand_Execute_UnknownSubcommand(t *testing.T) {
	args := []string{"invalid"}
	mockKademlia := newMockKademlia(
		t,
		func() []*kademlia.Bucket {
			t.Fatalf("Execute(%v) called GetBuckets", args)
			return nil
		},
		func() []string {
			t.Fatalf("Execute(%v) called GetStoredKeys", args)
			return nil
		},
	)
	var mockOut bytes.Buffer
	command := show.NewShowCommand(mockKademlia, &mockOut)
	err := command.Execute(args)
	if err == nil || !strings.Contains(err.Error(), "unknown command") {
		t.Fatalf("Execute(%v) err = %v; want unknown command", args, err)
	}
}

func TestShowCommand_Execute_RT_TooManyArgs(t *testing.T) {
	args := []string{"rt", "extra"}
	mockKademlia := newMockKademlia(
		t,
		func() []*kademlia.Bucket {
			t.Fatalf("Execute(%v) called GetBuckets", args)
			return nil
		},
		func() []string {
			t.Fatalf("Execute(%v) called GetStoredKeys", args)
			return nil
		},
	)
	var mockOut bytes.Buffer
	command := show.NewShowCommand(mockKademlia, &mockOut)
	err := command.Execute(args)
	if err == nil {
		t.Fatalf("Execute(%v) expected error, got nil", args)
	}
}

func TestShowCommand_Execute_RT_Empty(t *testing.T) {
	getBucketsCalled := false
	args := []string{"rt"}
	mockKademlia := newMockKademlia(
		t,
		func() []*kademlia.Bucket {
			getBucketsCalled = true
			return []*kademlia.Bucket{}
		},
		func() []string {
			t.Fatalf("Execute(%v) called GetStoredKeys", args)
			return nil
		},
	)
	var mockOut bytes.Buffer
	command := show.NewShowCommand(mockKademlia, &mockOut)
	err := command.Execute(args)
	if err != nil {
		t.Fatalf("Execute(%v) returned unexpected error: %v", args, err)
	}
	if !getBucketsCalled {
		t.Fatalf("Execute(%v) didn't call GetBuckets", args)
	}
	output := mockOut.String()
	if output != "Routing table is empty.\n" {
		t.Fatalf(
			"Execute(%v) expected empty table output but got\n%s",
			args,
			output,
		)
	}
}

func TestShowCommand_Execute_RT_Foo(t *testing.T) {
	getBucketsCalled := false
	args := []string{"rt"}
	mockKademlia := newMockKademlia(
		t,
		func() []*kademlia.Bucket {
			getBucketsCalled = true
			bucket1 := kademlia.NewBucket()
			bucket1.AddContact(kademlia.NewContactFromAddress(
				entities.Address{IP: "127.0.0.1", Port: 8080},
			))
			bucket1.AddContact(kademlia.NewContactFromAddress(
				entities.Address{IP: "127.0.0.1", Port: 8081},
			))
			bucket2 := kademlia.NewBucket()
			bucket3 := kademlia.NewBucket()
			bucket3.AddContact(kademlia.NewContactFromAddress(
				entities.Address{IP: "127.0.0.1", Port: 8082},
			))
			return []*kademlia.Bucket{
				bucket1,
				bucket2,
				bucket3,
			}
		},
		func() []string {
			t.Fatalf("Execute(%v) called GetStoredKeys", args)
			return nil
		},
	)
	var mockOut bytes.Buffer
	command := show.NewShowCommand(mockKademlia, &mockOut)
	err := command.Execute(args)
	if err != nil {
		t.Fatalf("Execute(%v) returned unexpected error: %v", args, err)
	}
	if !getBucketsCalled {
		t.Fatalf("Execute(%v) didn't call GetBuckets", args)
	}
	output := mockOut.String()
	expectedOutput := "Bucket 0:\n" +
		"    -  0 5a75d60e...e89e5c9f 127.0.0.1:8081\n" +
		"    -  1 7b971d01...7f49a7c6 127.0.0.1:8080\n" +
		"Bucket 2:\n" +
		"    -  0 44578d7c...65db428e 127.0.0.1:8082\n"
	if output != expectedOutput {
		t.Fatalf(
			"Execute(%v) invalid output, got\n%s",
			args,
			output,
		)
	}
}

func TestShowCommand_Execute_DS_TooManyArgs(t *testing.T) {
	args := []string{"ds", "extra"}
	mockKademlia := newMockKademlia(
		t,
		func() []*kademlia.Bucket {
			t.Fatalf("Execute(%v) called GetBuckets", args)
			return nil
		},
		func() []string {
			t.Fatalf("Execute(%v) called GetStoredKeys", args)
			return nil
		},
	)
	var mockOut bytes.Buffer
	command := show.NewShowCommand(mockKademlia, &mockOut)
	err := command.Execute(args)
	if err == nil {
		t.Fatalf("Execute(%v) expected error, got nil", args)
	}
}

func TestShowCommand_Execute_DS_Empty(t *testing.T) {
	getStoredKeysCalled := false
	args := []string{"ds"}
	mockKademlia := newMockKademlia(
		t,
		func() []*kademlia.Bucket {
			t.Fatalf("Execute(%v) called GetBuckets", args)
			return nil
		},
		func() []string {
			getStoredKeysCalled = true
			return []string{}
		},
	)
	var mockOut bytes.Buffer
	command := show.NewShowCommand(mockKademlia, &mockOut)
	err := command.Execute(args)
	if err != nil {
		t.Fatalf("Execute(%v) returned unexpected error: %v", args, err)
	}
	if !getStoredKeysCalled {
		t.Fatalf("Execute(%v) didn't call GetStoredKeys", args)
	}
	output := mockOut.String()
	if output != "" {
		t.Fatalf("Execute(%v) exepted empty output but got\n%s", args, output)
	}
}

func TestShowCommand_Execute_DS_WithKeys(t *testing.T) {
	getStoredKeysCalled := false
	args := []string{"ds"}
	mockKademlia := newMockKademlia(
		t,
		func() []*kademlia.Bucket {
			t.Fatalf("Execute(%v) called GetBuckets", args)
			return nil
		},
		func() []string {
			getStoredKeysCalled = true
			return []string{"key1", "key2"}
		},
	)
	var mockOut bytes.Buffer
	command := show.NewShowCommand(mockKademlia, &mockOut)
	err := command.Execute(args)
	if err != nil {
		t.Fatalf("Execute(%v) returned unexpected error: %v", args, err)
	}
	if !getStoredKeysCalled {
		t.Fatalf("Execute(%v) didn't call GetStoredKeys", args)
	}
	output := mockOut.String()
	if output != "key1\nkey2\n" {
		t.Fatalf(
			"Execute(%v) exepted keys listed output but got\n%s",
			args,
			output,
		)
	}
}

func TestShowCommand_GetCompletions(t *testing.T) {
	mockKademlia := newMockKademlia(
		t,
		func() []*kademlia.Bucket {
			t.Fatal("GetCompletions() called GetBuckets")
			return nil
		},
		func() []string {
			t.Fatal("GetCompletions() called GetStoredKeys")
			return nil
		},
	)
	var mockOut bytes.Buffer
	command := show.NewShowCommand(mockKademlia, &mockOut)
	completions := command.GetCompletions()
	expectedCompletions := []commands.Completion{
		{
			Name: "ds",
			Children: []commands.Completion{
				{Name: "--help"},
				{Name: "-h"},
			},
		},
		{
			Name: "rt",
			Children: []commands.Completion{
				{Name: "--help"},
				{Name: "-h"},
			},
		},
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
