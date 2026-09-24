package commands_test

import (
	"errors"
	"kademlia/config"
	"kademlia/internal/core/entities"
	"kademlia/internal/core/kademlia"
	"kademlia/internal/shell/commands"
	"reflect"
	"strings"
	"testing"
	"time"
)

func NewMockKademlia(
	t *testing.T,
	pingFunction func(address entities.Address) (time.Duration, error),
) kademlia.Kademlia {
	mockKademlia := &kademlia.MockKademlia{
		RunFunction: func() error {
			t.Fatal("Run should not be called")
			return nil
		},
		PingFunction: pingFunction,
	}
	return mockKademlia
}

func TestNewPingCommand(t *testing.T) {
	mockKademlia := NewMockKademlia(
		t,
		func(address entities.Address) (time.Duration, error) {
			t.Fatal("NewPingCommand() called Ping")
			return 0, nil
		},
	)
	command := commands.NewPingCommand(mockKademlia)
	if command == nil {
		t.Fatalf("NewPingCommand() returned nil")
	}
}

func TestPingCommand_Execute_NoArgs(t *testing.T) {
	args := []string{}
	kademlia := NewMockKademlia(
		t,
		func(address entities.Address) (time.Duration, error) {
			t.Fatalf("Execute(%v) called Ping", args)
			return 0, nil
		},
	)
	command := commands.NewPingCommand(kademlia)
	err := command.Execute(args)
	if err == nil {
		t.Fatalf("Execute(%v) expected error, got nil", args)
	}
}

func TestPingCommand_Execute_TooManyArgs(t *testing.T) {
	args := []string{"127.0.0.1", "8080", "extra"}
	mockKademlia := NewMockKademlia(
		t,
		func(address entities.Address) (time.Duration, error) {
			t.Fatalf("Execute(%v) called Ping", args)
			return 0, nil
		},
	)
	command := commands.NewPingCommand(mockKademlia)
	err := command.Execute(args)
	expectedError := "accepts between 1 and 2 arg(s), received 3"
	if err == nil || err.Error() != expectedError {
		t.Fatalf("Execute(%v) err = %v; want %v", args, err, expectedError)
	}
}

func TestPingCommand_Execute_AddressAndPort(t *testing.T) {
	pingCalled := false
	var parsedAddress entities.Address
	mockKademlia := NewMockKademlia(
		t,
		func(address entities.Address) (time.Duration, error) {
			pingCalled = true
			parsedAddress = address
			return 0, nil
		},
	)
	command := commands.NewPingCommand(mockKademlia)
	args := []string{"127.0.0.1", "8081"}
	err := command.Execute(args)
	if err != nil {
		t.Fatalf("Execute(%v) returned unexpected error: %v", args, err)
	}
	if !pingCalled {
		t.Fatalf("Execute(%v) didn't call Ping", args)
	}
	expectedIP := "127.0.0.1"
	if parsedAddress.IP != expectedIP {
		t.Fatalf("unexpected IP: got %q, want %q", parsedAddress.IP, expectedIP)
	}
	expectedPort := 8081
	if parsedAddress.Port != expectedPort {
		t.Fatalf(
			"unexpected port: got %d, want %d",
			parsedAddress.Port,
			expectedPort,
		)
	}
}

func TestPingCommand_Execute_AddressOnly_UsesDefaultPort(t *testing.T) {
	pingCalled := false
	var parsedAddress entities.Address
	mockKademlia := NewMockKademlia(
		t,
		func(address entities.Address) (time.Duration, error) {
			pingCalled = true
			parsedAddress = address
			return 0, nil
		},
	)
	command := commands.NewPingCommand(mockKademlia)
	args := []string{"127.0.0.1"}
	err := command.Execute(args)
	if err != nil {
		t.Fatalf("Execute(%v) returned unexpected error: %v", args, err)
	}
	if !pingCalled {
		t.Fatalf("Execute(%v) didn't call Ping", args)
	}
	if parsedAddress.Port != config.DefaultPort {
		t.Fatalf(
			"unexpected port: got %d, want %d",
			parsedAddress.Port,
			config.DefaultPort,
		)
	}
}

func TestPingCommand_Execute_InvalidPort(t *testing.T) {
	args := []string{"127.0.0.1", "not-a-port"}
	mockKademlia := NewMockKademlia(
		t,
		func(address entities.Address) (time.Duration, error) {
			t.Fatalf("Execute(%v) called Ping", args)
			return 0, nil
		},
	)
	command := commands.NewPingCommand(mockKademlia)
	err := command.Execute(args)
	if err == nil {
		t.Fatalf("Execute(%v) expected error, got nil", args)
	}
	if !strings.Contains(err.Error(), "invalid port") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestPingCommand_Execute_PingFail(t *testing.T) {
	pingCalled := false
	args := []string{"127.0.0.1"}
	mockKademlia := NewMockKademlia(
		t,
		func(address entities.Address) (time.Duration, error) {
			pingCalled = true
			return 0, errors.New("test error")
		},
	)
	command := commands.NewPingCommand(mockKademlia)
	err := command.Execute(args)
	if err != nil {
		t.Fatalf("Execute(%v) returned unexpected error: %v", args, err)
	}
	if !pingCalled {
		t.Fatalf("Execute(%v) didn't call Ping", args)
	}
}

func TestPingCommand_GetCompletions(t *testing.T) {
	mockKademlia := NewMockKademlia(
		t,
		func(address entities.Address) (time.Duration, error) {
			t.Fatal("GetCompletions() called Ping")
			return 0, nil
		},
	)
	command := commands.NewPingCommand(mockKademlia)
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
