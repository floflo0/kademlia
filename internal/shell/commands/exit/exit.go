package exit

import (
	"kademlia/internal/core/kademlia"
	"kademlia/internal/shell/commands"
	"log/slog"

	"github.com/spf13/cobra"
)

type exitCommand struct {
	kademlia kademlia.Kademlia
}

func NewExitCommand(kademlia kademlia.Kademlia) commands.Command {
	return &exitCommand{
		kademlia: kademlia,
	}
}

func (c *exitCommand) buildCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "exit [-h]",
		Short:                 "Exit Kademlia",
		Args:                  cobra.NoArgs,
		DisableFlagsInUseLine: true,
		RunE: func(command *cobra.Command, args []string) error {
			slog.Debug("Running exit command")
			err := c.kademlia.Quit()
			if err != nil {
				return err
			}
			return nil
		},
	}
	return cmd
}

func (c *exitCommand) Execute(args []string) error {
	command := c.buildCommand()
	command.SetArgs(args)
	return command.Execute()
}

func (c *exitCommand) GetCompletions() []commands.Completion {
	return commands.GetCompletions(c.buildCommand())
}
