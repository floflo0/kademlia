package get

import (
	"fmt"
	"io"
	"kademlia/internal/core/kademlia"
	"kademlia/internal/shell/commands"
	"log/slog"

	"github.com/spf13/cobra"
)

type getCommand struct {
	kademlia kademlia.Kademlia
	out      io.Writer
}

func NewGetCommand(kademlia kademlia.Kademlia, out io.Writer) commands.Command {
	return &getCommand{
		kademlia: kademlia,
		out:      out,
	}
}

func (c *getCommand) buildCommand() *cobra.Command {
	command := &cobra.Command{
		Use:                   "get [-h]",
		Short:                 "Get a value from a key",
		Args:                  cobra.NoArgs,
		Example:               "get KEY [FILENAME]",
		DisableFlagsInUseLine: true,
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				panic("unreachable")
			}
			return fmt.Errorf("missing subcommand")
		},
	}
	command.AddCommand(
		getCommandFunc(c.kademlia, c.out),
	)
	command.SetOut(c.out)
	return command
}

func (c *getCommand) Execute(args []string) error {
	command := c.buildCommand()
	return command.Execute()
}

func (c *getCommand) GetCompletions() []commands.Completion {
	return commands.GetCompletions(c.buildCommand())
}

func getCommandFunc(
	kademlia kademlia.Kademlia,
	out io.Writer,
) *cobra.Command {
	return &cobra.Command{
		Use:                   "[-h]",
		Short:                 "",
		Args:                  cobra.NoArgs,
		DisableFlagsInUseLine: true,
		RunE: func(command *cobra.Command, args []string) error {
			slog.Debug("Running get command")
			value, contactID, err := kademlia.GetValue(args[0])
			if err != nil {
				return err
			}
			if value != nil {
				slog.Info("Value found")
				if len(args) == 1 {
					fmt.Fprintf(out, "Value for key %v has been found at :%v", args[0], contactID)
					return nil
				}
				if len(args) == 2 {
					slog.Info("Have to download")
					return nil
				}
			}
			fmt.Fprintf(out, "Value for key %v not Found", args[0])
			return nil
		},
	}
}
