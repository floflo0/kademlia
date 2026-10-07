package publish

import (
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"kademlia/internal/core/kademlia"
	"kademlia/internal/shell/commands"

	"github.com/spf13/cobra"
)

type publishCommand struct {
	kademlia kademlia.Kademlia
	dns      kademlia.DNSVerifier
	out      io.Writer
}

func NewPublishCommand(kademlia kademlia.Kademlia, dns kademlia.DNSVerifier, out io.Writer) commands.Command {
	return &publishCommand{
		kademlia: kademlia,
		dns:      dns,
		out:      out,
	}
}

func (c *publishCommand) buildCommand() *cobra.Command {
	var (
		domain   string
		pkgName  string
		version  string
		filePath string
		privHex  string
		force    bool
		prev     string
	)

	command := &cobra.Command{
		Use:   "publish [--force] [--prev=PREVIOUS-VERSION] DOMAIN:PACKAGE:VERSION FILENAME",
		Short: "Publish a new package version to the DHT registry.",
		RunE: func(cmd *cobra.Command, args []string) error {
			// 1. Parse positional arguments (DOMAIN:PACKAGE:VERSION FILENAME)
			if len(args) >= 2 {
				parts := strings.Split(args[0], ":")
				if len(parts) == 3 {
					domain = parts[0]
					pkgName = parts[1]
					version = parts[2]
				} else {
					return fmt.Errorf("invalid package format '%s', expected DOMAIN:PACKAGE:VERSION", args[0])
				}
				filePath = args[1]
			}

			// 2. Fallback validation check
			if domain == "" || pkgName == "" || version == "" || filePath == "" {
				return fmt.Errorf("missing parameters. Usage: publish [--force] [--prev=VERSION] -k KEY DOMAIN:PACKAGE:VERSION FILENAME")
			}

			slog.Debug("Running publish command",
				"domain", domain,
				"package", pkgName,
				"version", version,
				"force", force,
				"prev", prev,
			)

			// 3. Read payload file
			fileContent, err := os.ReadFile(filePath)
			if err != nil {
				return fmt.Errorf("failed to read file '%s': %w", filePath, err)
			}

			// 4. Decode Ed25519 private key
			privKeyBytes, err := hex.DecodeString(privHex)
			if err != nil || len(privKeyBytes) != ed25519.PrivateKeySize {
				return fmt.Errorf("invalid ed25519 private key hex")
			}
			privKey := ed25519.PrivateKey(privKeyBytes)

			// 5. Invoke package service with force and prev parameters
			record, err := c.kademlia.PublishPackage(domain, pkgName, version, string(fileContent), privKey, c.dns, force, prev)
			if err != nil {
				return err
			}

			recHash, _ := record.Hash()
			fmt.Fprintf(c.out, "Package successfully published!\nVersion Record Hash: %s\n", recHash)
			return nil
		},
	}

	// Flags definition
	command.Flags().BoolVar(&force, "force", false, "Bypass version sequence validation checks")
	command.Flags().StringVar(&prev, "prev", "", "Explicitly specify the previous version")
	command.Flags().StringVarP(&domain, "domain", "d", "", "Domain owner name (optional if positional argument is used)")
	command.Flags().StringVarP(&pkgName, "package", "p", "", "Package name (optional if positional argument is used)")
	command.Flags().StringVarP(&version, "version", "v", "", "Version string (optional if positional argument is used)")
	command.Flags().StringVarP(&filePath, "file", "f", "", "File path to payload (optional if positional argument is used)")
	command.Flags().StringVarP(&privHex, "key", "k", "", "Ed25519 private key in hex format")

	_ = command.MarkFlagRequired("key")

	command.SetOut(c.out)
	return command
}

func (c *publishCommand) Execute(args []string) error {
	command := c.buildCommand()
	command.SetArgs(args)
	return command.Execute()
}

func (c *publishCommand) GetCompletions() []commands.Completion {
	return commands.GetCompletions(c.buildCommand())
}
