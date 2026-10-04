package install

import (
	"fmt"
	"io"
	"kademlia/internal/core/kademlia"
	"kademlia/internal/shell/commands"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

type installCommand struct {
	kademlia kademlia.Kademlia
	out      io.Writer
}

func NewInstallCommand(kademlia kademlia.Kademlia, out io.Writer) commands.Command {
	return &installCommand{
		kademlia: kademlia,
		out:      out,
	}
}

func (c *installCommand) buildCommand() *cobra.Command {
	var (
		domain     string
		pkgName    string
		version    string
		outputPath string
	)

	command := &cobra.Command{
		Use:   "install -d domain -p package [-v version] [-o output_file]",
		Short: "Fetch and install a package version from the DHT registry.",
		RunE: func(cmd *cobra.Command, args []string) error {
			slog.Debug("Running install command", "domain", domain, "package", pkgName, "version", version)

			blobContent, resolvedVersion, err := c.kademlia.InstallPackage(domain, pkgName, version)
			if err != nil {
				return err
			}

			if outputPath != "" {
				if err := os.WriteFile(outputPath, []byte(blobContent), 0644); err != nil {
					return fmt.Errorf("failed to save output file: %w", err)
				}
				fmt.Fprintf(c.out, "Successfully installed %s/%s@%s to %s\n", domain, pkgName, resolvedVersion, outputPath)
			} else {
				fmt.Fprintf(c.out, "Successfully installed %s/%s@%s\nContent:\n%s\n", domain, pkgName, resolvedVersion, blobContent)
			}

			return nil
		},
	}

	command.Flags().StringVarP(&domain, "domain", "d", "", "Domain owner name")
	command.Flags().StringVarP(&pkgName, "package", "p", "", "Package name")
	command.Flags().StringVarP(&version, "version", "v", "latest", "Version string (default: latest)")
	command.Flags().StringVarP(&outputPath, "out", "o", "", "File path to save the installed binary")

	_ = command.MarkFlagRequired("domain")
	_ = command.MarkFlagRequired("package")

	command.SetOut(c.out)
	return command
}

func (c *installCommand) Execute(args []string) error {
	command := c.buildCommand()
	command.SetArgs(args)
	return command.Execute()
}

func (c *installCommand) GetCompletions() []commands.Completion {
	return commands.GetCompletions(c.buildCommand())
}
