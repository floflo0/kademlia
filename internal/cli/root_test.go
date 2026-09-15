package cli_test

import (
	"bytes"
	"errors"
	"kademlia/config"
	"kademlia/internal/cli"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestNewRootCommand(t *testing.T) {
	commandName := "kademlia"
	rootCommand := cli.NewRootCommand(
		commandName,
		func(appConfig cli.Config) error {
			t.Fatalf("NewRootCommand(%q) called start function", commandName)
			return nil
		},
	)
	if rootCommand == nil {
		t.Fatalf("NewRootCommand(%q) returned nil", commandName)
	}
}

func mockRootCommand(
	start func(config cli.Config) error,
	args []string,
) (*cobra.Command, *bytes.Buffer, *bytes.Buffer) {
	rootCommand := cli.NewRootCommand("kademlia", start)
	rootCommand.SetArgs(args)

	var mockOut bytes.Buffer
	rootCommand.SetOut(&mockOut)
	var mockErr bytes.Buffer
	rootCommand.SetErr(&mockErr)

	return rootCommand, &mockOut, &mockErr
}

func TestNewRootCommand_Execute_NoArgs(t *testing.T) {
	startCalled := false
	rootCommand, _, _ := mockRootCommand(
		func(appConfig cli.Config) error {
			if appConfig.Host != config.DefaultHost {
				t.Fatalf(
					"unexpected host: got %q, want %q",
					appConfig.Host,
					config.DefaultHost,
				)
			}
			if appConfig.Port != config.DefaultPort {
				t.Fatalf(
					"unexpected port: got %d, want %d",
					appConfig.Port,
					config.DefaultPort,
				)
			}
			startCalled = true
			return nil
		},
		[]string{},
	)
	err := rootCommand.Execute()
	if err != nil {
		t.Fatalf("Execute() returned unexpected error: %v", err)
	}
	if !startCalled {
		t.Fatalf("Execute() hasn't called start function")
	}
}

func TestNewRootCommand_Execute_StartReturnError(t *testing.T) {
	startCalled := false
	errorMessage := "test error"
	testError := errors.New(errorMessage)
	rootCommand, _, mockErr := mockRootCommand(
		func(appConfig cli.Config) error {
			startCalled = true
			return testError
		},
		[]string{},
	)
	err := rootCommand.Execute()
	if !errors.Is(err, testError) {
		t.Fatalf("Execute() err = %v; want %v", err, testError)
	}

	output := mockErr.String()
	if !strings.Contains(output, errorMessage) {
		t.Fatalf("expected error message output but got\n%s", output)
	}

	if !startCalled {
		t.Fatalf("Execute() hasn't called start function")
	}
}

func TestNewRootCommand_Execute_ShortHelpFlag(t *testing.T) {
	rootCommand, mockOut, _ := mockRootCommand(
		func(appConfig cli.Config) error {
			t.Fatalf("Execute() called start function")
			return nil
		},
		[]string{"-h"},
	)
	err := rootCommand.Execute()
	if err != nil {
		t.Fatalf("Execute() returned unexpected error: %v", err)
	}
	output := mockOut.String()
	if !strings.Contains(output, "Usage") {
		t.Fatalf("expected help output but got\n%s", output)
	}
}

func TestNewRootCommand_Execute_LongHelpFlag(t *testing.T) {
	rootCommand, mockOut, _ := mockRootCommand(
		func(appConfig cli.Config) error {
			t.Fatalf("Execute() called start function")
			return nil
		},
		[]string{"--help"},
	)
	err := rootCommand.Execute()
	if err != nil {
		t.Fatalf("Execute() returned unexpected error: %v", err)
	}
	output := mockOut.String()
	if !strings.Contains(output, "Usage") {
		t.Fatalf("expected help output but got\n%s", output)
	}
}

func TestNewRootCommand_Execute_HostFlag(t *testing.T) {
	startCalled := false
	rootCommand, _, _ := mockRootCommand(
		func(appConfig cli.Config) error {
			host := "0.0.0.0"
			if appConfig.Host != host {
				t.Fatalf(
					"unexpected host: got %q, want %q",
					appConfig.Host,
					host,
				)
			}
			startCalled = true
			return nil
		},
		[]string{"--host", "0.0.0.0"},
	)
	err := rootCommand.Execute()
	if err != nil {
		t.Fatalf("Execute() returned unexpected error: %v", err)
	}
	if !startCalled {
		t.Fatalf("Execute() hasn't called start function")
	}
}

func TestNewRootCommand_Execute_ShortPortFlag(t *testing.T) {
	startCalled := false
	rootCommand, _, _ := mockRootCommand(
		func(appConfig cli.Config) error {
			port := 8081
			if appConfig.Port != port {
				t.Fatalf(
					"unexpected port: got %d, want %d",
					appConfig.Port,
					port,
				)
			}
			startCalled = true
			return nil
		},
		[]string{"-p", "8081"},
	)
	err := rootCommand.Execute()
	if err != nil {
		t.Fatalf("Execute() returned unexpected error: %v", err)
	}
	if !startCalled {
		t.Fatalf("Execute() hasn't called start function")
	}
}

func TestNewRootCommand_Execute_LongPortFlag(t *testing.T) {
	startCalled := false
	rootCommand, _, _ := mockRootCommand(
		func(appConfig cli.Config) error {
			port := 8081
			if appConfig.Port != port {
				t.Fatalf(
					"unexpected port: got %d, want %d",
					appConfig.Port,
					port,
				)
			}
			startCalled = true
			return nil
		},
		[]string{"--port", "8081"},
	)
	err := rootCommand.Execute()
	if err != nil {
		t.Fatalf("Execute() returned unexpected error: %v", err)
	}
	if !startCalled {
		t.Fatalf("Execute() hasn't called start function")
	}
}
