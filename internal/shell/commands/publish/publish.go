package publish

import (
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"io"
	"kademlia/internal/core/kademlia"
	"kademlia/internal/shell/commands"
	"log/slog"
	"os"

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
	)

	command := &cobra.Command{
		Use:   "publish -d domain -p package -v version -f file -k key",
		Short: "Publish a new package version to the DHT registry.",
		RunE: func(cmd *cobra.Command, args []string) error {
			slog.Debug("Running publish command", "domain", domain, "package", pkgName, "version", version)

			fileContent, err := os.ReadFile(filePath)
			if err != nil {
				return err
			}

			privKeyBytes, err := hex.DecodeString(privHex)
			if err != nil || len(privKeyBytes) != ed25519.PrivateKeySize {
				return fmt.Errorf("invalid ed25519 private key hex")
			}
			privKey := ed25519.PrivateKey(privKeyBytes)

			record, err := c.kademlia.PublishPackage(domain, pkgName, version, string(fileContent), privKey, c.dns)
			if err != nil {
				return err
			}

			recHash, _ := record.Hash()
			fmt.Fprintf(c.out, "Package successfully published!\nVersion Record Hash: %s\n", recHash)
			return nil
		},
	}

	command.Flags().StringVarP(&domain, "domain", "d", "", "Domain owner name")
	command.Flags().StringVarP(&pkgName, "package", "p", "", "Package name")
	command.Flags().StringVarP(&version, "version", "v", "", "Version string (e.g. 1.0.0)")
	command.Flags().StringVarP(&filePath, "file", "f", "", "File path to binary payload")
	command.Flags().StringVarP(&privHex, "key", "k", "", "Ed25519 private key in hex format")

	_ = command.MarkFlagRequired("domain")
	_ = command.MarkFlagRequired("package")
	_ = command.MarkFlagRequired("version")
	_ = command.MarkFlagRequired("file")
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
