package cli_test

import (
	"bytes"
	"kademlia/internal/cli"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestNewRootCmd(t *testing.T) {
	commandName := "kademlia"
	rootCmd := cli.NewRootCmd(commandName)
	if rootCmd == nil {
		t.Fatalf("NewRootCmd(%q) returned nil", commandName)
	}
}

func mockRootCmd(args []string) (*cobra.Command, *bytes.Buffer) {
	rootCmd := cli.NewRootCmd("kademlia")
	rootCmd.SetArgs(args)

	var mockOut bytes.Buffer
	rootCmd.SetOut(&mockOut)

	return rootCmd, &mockOut
}

func TestNewRootCmd_Execute_NoArgs(t *testing.T) {
	rootCmd, _ := mockRootCmd([]string{})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Execute() returned unexpected error: %v", err)
	}
}

func TestNewRootCmd_Execute_ShortHelpFlag(t *testing.T) {
	rootCmd, mockOut := mockRootCmd([]string{"-h"})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Execute() returned unexpected error: %v", err)
	}
	output := mockOut.String()
	if !strings.Contains(output, "Usage") {
		t.Fatalf("expected help output but got\n%s", output)
	}
}

func TestNewRootCmd_Execute_LongHelpFlag(t *testing.T) {
	rootCmd, mockOut := mockRootCmd([]string{"--help"})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Execute() returned unexpected error: %v", err)
	}
	output := mockOut.String()
	if !strings.Contains(output, "Usage") {
		t.Fatalf("expected help output but got\n%s", output)
	}
}

func TestNewRootCmd_Execute_HostFlag(t *testing.T) {
	rootCmd, _ := mockRootCmd([]string{"--host", "0.0.0.0"})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Execute() returned unexpected error: %v", err)
	}
}

func TestNewRootCmd_Execute_ShortPortFlag(t *testing.T) {
	rootCmd, _ := mockRootCmd([]string{"-p", "8081"})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Execute() returned unexpected error: %v", err)
	}
}

func TestNewRootCmd_Execute_LongPortFlag(t *testing.T) {
	rootCmd, _ := mockRootCmd([]string{"--port", "8081"})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Execute() returned unexpected error: %v", err)
	}
}
