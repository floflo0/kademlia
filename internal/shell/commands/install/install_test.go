package install_test

import (
	"bytes"
	"crypto/ed25519"
	"errors"
	"fmt"
	"kademlia/internal/core/entities"
	"kademlia/internal/core/kademlia"
	"kademlia/internal/shell/commands/install"
	"os"
	"path/filepath"
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
		func(value string) (*entities.KademliaID, error) {
			t.Fatal("Put should not be called")
			return nil, nil
		},
		func() []*kademlia.Bucket {
			t.Fatal("GetBucketsFunction should not be called")
			return []*kademlia.Bucket{}
		},
		func() []entities.KademliaID {
			t.Fatal("GetStoredKeysFunction should not be called")
			return []entities.KademliaID{}
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

type mockKademliaWithInstall struct {
	kademlia.Kademlia
	installFunc func(domain, packageName, version string) (string, string, error)
}

func (m *mockKademliaWithInstall) PublishPackage(domain, packageName, version, blob string, privKey ed25519.PrivateKey, dns kademlia.DNSVerifier) (*kademlia.VersionRecord, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockKademliaWithInstall) InstallPackage(domain, packageName, version string) (string, string, error) {
	if m.installFunc != nil {
		return m.installFunc(domain, packageName, version)
	}
	return "", "", fmt.Errorf("not implemented")
}

func TestInstallCommand_Execute_MissingFlags(t *testing.T) {
	mockK := newMockKademlia(t)
	var mockOut bytes.Buffer

	cmd := install.NewInstallCommand(mockK, &mockOut)
	err := cmd.Execute([]string{})
	if err == nil {
		t.Fatal("Expected error for missing required flags, got nil")
	}
}

func TestInstallCommand_Execute_Success_ConsoleOutput(t *testing.T) {
	mockK := newMockKademlia(t)
	mockKWithInstall := &mockKademliaWithInstall{
		Kademlia: mockK,
		installFunc: func(domain, packageName, version string) (string, string, error) {
			return "binary_data_here", "1.0.0", nil
		},
	}

	var mockOut bytes.Buffer
	cmd := install.NewInstallCommand(mockKWithInstall, &mockOut)

	args := []string{"-d", "example.com", "-p", "my-lib", "-v", "1.0.0"}
	err := cmd.Execute(args)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := "Successfully installed example.com/my-lib@1.0.0\nContent:\nbinary_data_here\n"
	if mockOut.String() != expected {
		t.Fatalf("Expected output %q, got %q", expected, mockOut.String())
	}
}

func TestInstallCommand_Execute_Success_FileOutput(t *testing.T) {
	outDir := t.TempDir()
	outFile := filepath.Join(outDir, "installed.bin")

	mockK := newMockKademlia(t)
	mockKWithInstall := &mockKademliaWithInstall{
		Kademlia: mockK,
		installFunc: func(domain, packageName, version string) (string, string, error) {
			return "binary_data_here", "1.0.0", nil
		},
	}

	var mockOut bytes.Buffer
	cmd := install.NewInstallCommand(mockKWithInstall, &mockOut)

	args := []string{"-d", "example.com", "-p", "my-lib", "-o", outFile}
	err := cmd.Execute(args)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	data, err := os.ReadFile(outFile)
	if err != nil || string(data) != "binary_data_here" {
		t.Fatalf("File content mismatch or read error: %v", err)
	}
}

func TestInstallCommand_Execute_PackageNotFound(t *testing.T) {
	mockK := newMockKademlia(t)
	expectedErr := errors.New("package not found")
	mockKWithInstall := &mockKademliaWithInstall{
		Kademlia: mockK,
		installFunc: func(domain, packageName, version string) (string, string, error) {
			return "", "", expectedErr
		},
	}

	var mockOut bytes.Buffer
	cmd := install.NewInstallCommand(mockKWithInstall, &mockOut)

	args := []string{"-d", "example.com", "-p", "nonexistent"}
	err := cmd.Execute(args)
	if !errors.Is(err, expectedErr) {
		t.Fatalf("Expected error %v, got %v", expectedErr, err)
	}
}
