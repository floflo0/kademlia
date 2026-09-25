package show

import (
	"fmt"
	"io"
	"kademlia/internal/core/kademlia"
	"kademlia/internal/shell/commands"
	"log/slog"

	"github.com/spf13/cobra"
)

type showCommand struct {
	kademlia kademlia.Kademlia
	out      io.Writer
}

func NewShowCommand(kademlia kademlia.Kademlia, out io.Writer) commands.Command {
	return &showCommand{
		kademlia: kademlia,
		out:      out,
	}
}

func (c *showCommand) buildCommand() *cobra.Command {
	command := &cobra.Command{
		Use:                   "show [-h] rt|ds",
		Short:                 "Show debug information",
		Args:                  cobra.NoArgs,
		Example:               "show rt",
		DisableFlagsInUseLine: true,
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 0 {
				panic("unreachable")
			}
			return fmt.Errorf("missing subcommand")
		},
	}
	command.AddCommand(
		showRoutingTableCommand(c.kademlia, c.out),
		showDataStoreCommand(c.kademlia, c.out),
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

func showRoutingTableCommand(
	kademlia kademlia.Kademlia,
	out io.Writer,
) *cobra.Command {
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

func showDataStoreCommand(
	kademlia kademlia.Kademlia,
	out io.Writer,
) *cobra.Command {
	return &cobra.Command{
		Use:                   "ds [-h]",
		Short:                 "Show the data store keys",
		Args:                  cobra.NoArgs,
		DisableFlagsInUseLine: true,
		RunE: func(command *cobra.Command, args []string) error {
			slog.Debug("Running show ds command")
			for _, key := range kademlia.GetStoredKeys() {
				fmt.Fprintln(out, key)
			}
			return nil
		},
	}
}
