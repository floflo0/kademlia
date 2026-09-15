package cli_test

import (
	"bytes"
	"kademlia/internal/cli"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

type mockShell struct {
	called bool
}

func newMockShell() *mockShell {
	return &mockShell{
		called: false,
	}
}

func (s *mockShell) Run() error {
	s.called = true
	return nil
}

func TestNewRootCmd(t *testing.T) {
	commandName := "kademlia"
	shell := newMockShell()
	rootCmd := cli.NewRootCmd(commandName, shell)
	if rootCmd == nil {
		t.Fatalf("NewRootCmd(%q) returned nil", commandName)
	}
	if shell.called {
		t.Fatalf("NewRootCmd(%q) called shell.Run()", commandName)
	}
}

func mockRootCmd(args []string) (*cobra.Command, *mockShell, *bytes.Buffer) {
	shell := newMockShell()
	rootCmd := cli.NewRootCmd("kademlia", shell)
	rootCmd.SetArgs(args)

	var mockOut bytes.Buffer
	rootCmd.SetOut(&mockOut)

	return rootCmd, shell, &mockOut
}

func TestNewRootCmd_Execute_NoArgs(t *testing.T) {
	rootCmd, shell, _ := mockRootCmd([]string{})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Execute() returned unexpected error: %v", err)
	}
	if !shell.called {
		t.Fatalf("Execute() don't call shell.Run()")
	}
}

func TestNewRootCmd_Execute_ShortHelpFlag(t *testing.T) {
	rootCmd, shell, mockOut := mockRootCmd([]string{"-h"})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Execute() returned unexpected error: %v", err)
	}
	output := mockOut.String()
	if !strings.Contains(output, "Usage") {
		t.Fatalf("expected help output but got\n%s", output)
	}
	if shell.called {
		t.Fatalf("Execute() called shell.Run()")
	}
}

func TestNewRootCmd_Execute_LongHelpFlag(t *testing.T) {
	rootCmd, shell, mockOut := mockRootCmd([]string{"--help"})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Execute() returned unexpected error: %v", err)
	}
	output := mockOut.String()
	if !strings.Contains(output, "Usage") {
		t.Fatalf("expected help output but got\n%s", output)
	}
	if shell.called {
		t.Fatalf("Execute() called shell.Run()")
	}
}

func TestNewRootCmd_Execute_HostFlag(t *testing.T) {
	rootCmd, shell, _ := mockRootCmd([]string{"--host", "0.0.0.0"})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Execute() returned unexpected error: %v", err)
	}
	if !shell.called {
		t.Fatalf("Execute() don't call shell.Run()")
	}
}

func TestNewRootCmd_Execute_ShortPortFlag(t *testing.T) {
	rootCmd, shell, _ := mockRootCmd([]string{"-p", "8081"})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Execute() returned unexpected error: %v", err)
	}
	if !shell.called {
		t.Fatalf("Execute() don't call shell.Run()")
	}
}

func TestNewRootCmd_Execute_LongPortFlag(t *testing.T) {
	rootCmd, shell, _ := mockRootCmd([]string{"--port", "8081"})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Execute() returned unexpected error: %v", err)
	}
	if !shell.called {
		t.Fatalf("Execute() don't call shell.Run()")
	}
}
