package shell_test

import (
	"kademlia/internal/core/kademlia"
	"kademlia/internal/shell"
	"testing"
)

func TestNewShell(t *testing.T) {
	kademlia := &kademlia.MockKademlia{}
	shell := shell.NewShell(kademlia)
	if shell == nil {
		t.Fatalf("NewShell() returned nil")
	}
}

func TestShell_Run_Success(t *testing.T) {
	kademlia := &kademlia.MockKademlia{}
	shell := shell.NewShell(kademlia)

	err := shell.Run()
	if err != nil {
		t.Fatalf("Run() returned unexpected error: %v", err)
	}
}
