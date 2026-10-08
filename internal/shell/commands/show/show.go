package show

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

// DNSVerifier defines the interface required to resolve domain public keys locally.
type DNSVerifier interface {
	GetPublicKey(domain string) (ed25519.PublicKey, error)
}

type showCommand struct {
	kademlia kademlia.Kademlia
	dns      DNSVerifier
	out      io.Writer
}

// NewShowCommand constructs a show command with explicit typed parameters for compile-time type safety.
func NewShowCommand(kademlia kademlia.Kademlia, dns DNSVerifier, out io.Writer) commands.Command {
	if out == nil {
		out = os.Stdout
	}
	return &showCommand{
		kademlia: kademlia,
		dns:      dns,
		out:      out,
	}
}

func (c *showCommand) buildCommand() *cobra.Command {
	command := &cobra.Command{
		Use:                   "show [-h] rt|ds|dns|DOMAIN:PACKAGE",
		Short:                 "Show debug information, DNS info, or package version chain.",
		Args:                  cobra.ArbitraryArgs,
		DisableFlagsInUseLine: true,
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("missing subcommand")
			}

			// Handle 'show DOMAIN:PACKAGE' (Package version chain inspection)
			if len(args) == 1 && strings.Contains(args[0], ":") {
				parts := strings.Split(args[0], ":")
				if len(parts) == 2 {
					domain := parts[0]
					pkg := parts[1]

					chain, err := c.kademlia.GetVersionChain(domain, pkg)
					if err != nil {
						return fmt.Errorf("failed to fetch version chain: %w", err)
					}

					if len(chain) == 0 {
						fmt.Fprintf(c.out, "No version chain found for %s:%s\n", domain, pkg)
						return nil
					}

					fmt.Fprintf(c.out, "Version chain for %s:%s:\n", domain, pkg)
					for i, record := range chain {
						recHash, _ := record.Hash()
						fmt.Fprintf(c.out, "[%d] Version: %s | Hash: %s | PrevHash: %s\n",
							i+1, record.Version, recHash, record.PrevVersionRecordHash)
					}
					return nil
				}
			}

			return fmt.Errorf("unknown command %q for %q", args[0], cmd.CommandPath())
		},
	}

	command.AddCommand(
		showRoutingTableCommand(c.kademlia, c.out),
		showDataStoreCommand(c.kademlia, c.out),
		showDNSCommand(c.dns, c.out),
	)

	command.SetOut(c.out)
	return command
}

func (c *showCommand) Execute(args []string) error {
	command := c.buildCommand()
	command.SetArgs(args)
	return command.Execute()
}

func (c *showCommand) GetCompletions() []commands.Completion {
	return commands.GetCompletions(c.buildCommand())
}

func showRoutingTableCommand(kademlia kademlia.Kademlia, out io.Writer) *cobra.Command {
	return &cobra.Command{
		Use:                   "rt [-h]",
		Short:                 "Show the routing table",
		Args:                  cobra.NoArgs,
		DisableFlagsInUseLine: true,
		RunE: func(command *cobra.Command, args []string) error {
			slog.Debug("Running show rt command")
			routingTableEmpty := true
			for i, bucket := range kademlia.GetBuckets() {
				if bucket.Len() == 0 {
					continue
				}
				routingTableEmpty = false
				fmt.Fprintf(out, "Bucket %d:\n", i)
				for j, contact := range bucket.GetContacts() {
					fmt.Fprintf(
						out,
						"    - %2d %s %s\n",
						j,
						contact.ID.ShortString(),
						contact.Address.String(),
					)
				}
			}
			if routingTableEmpty {
				fmt.Fprintln(out, "Routing table is empty.")
				return nil
			}
			return nil
		},
	}
}

func showDataStoreCommand(kademlia kademlia.Kademlia, out io.Writer) *cobra.Command {
	return &cobra.Command{
		Use:                   "ds [-h]",
		Short:                 "Show the data store keys",
		Args:                  cobra.NoArgs,
		DisableFlagsInUseLine: true,
		RunE: func(command *cobra.Command, args []string) error {
			slog.Debug("Running show ds command")
			for _, key := range kademlia.GetStoredKeys() {
				fmt.Fprintln(out, key.String())
			}
			return nil
		},
	}
}

func showDNSCommand(dns DNSVerifier, out io.Writer) *cobra.Command {
	return &cobra.Command{
		Use:                   "dns DOMAIN",
		Short:                 "Show DNS public key for a domain",
		Args:                  cobra.ExactArgs(1),
		DisableFlagsInUseLine: true,
		RunE: func(command *cobra.Command, args []string) error {
			domain := args[0]
			slog.Debug("Running show dns command", "domain", domain)
			if dns == nil {
				return fmt.Errorf("DNS verifier not configured")
			}
			pubKey, err := dns.GetPublicKey(domain)
			if err != nil {
				return fmt.Errorf("failed to resolve DNS public key for %s: %w", domain, err)
			}
			fmt.Fprintf(out, "Domain: %s\nPublic Key (hex): %s\n", domain, hex.EncodeToString(pubKey))
			return nil
		},
	}
}
