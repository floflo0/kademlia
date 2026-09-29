package get

import (
	"errors"
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
		Use:                   "get [-h] key [filename]",
		Short:                 "Get a value from a key",
		Args:                  cobra.RangeArgs(1, 2),
		Example:               "get 083749...917e32 file.txt",
		DisableFlagsInUseLine: true,
		RunE: func(command *cobra.Command, args []string) error {
			slog.Debug("Running get command")
			if len(args[0]) != 64 {
				return errors.New("Bad length, must be 64")
			}
			value, contactID, err := c.kademlia.GetValue(args[0])
			if err != nil {
				if err == kademlia.ErrNoContact {
					return errors.New("Value not found")
				}
				return err
			}
			if value != nil {
				slog.Info("Value found")
				if len(args) == 1 {
					fmt.Fprintf(c.out, "Value for key %v has been found at :%v", args[0], contactID)
					return nil
				}
				if len(args) == 2 {
					slog.Info("Have to write in file")
					return nil
				}
			}
			fmt.Fprintf(c.out, "Value for key %v not Found", args[0])
			return nil
		},
	}
	command.SetOut(c.out)
	return command
}

func (c *getCommand) Execute(args []string) error {
	command := c.buildCommand()
	command.SetArgs(args)
	return command.Execute()
}

func (c *getCommand) GetCompletions() []commands.Completion {
	return commands.GetCompletions(c.buildCommand())
}
