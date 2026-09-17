package commands

import (
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

type exitCommand struct{}

func NewExitCommand() Command {
	return &exitCommand{}
}

func (*exitCommand) buildCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "exit [-h]",
		Short:                 "Exit Kademlia",
		Args:                  cobra.NoArgs,
		DisableFlagsInUseLine: true,
		Run:                   run,
	}
	return cmd
}

func (c *exitCommand) Execute(args []string) error {
	command := c.buildCommand()
	command.SetArgs(args)
	return command.Execute()
}

func (c *exitCommand) GetFlags() []string {
	return getFlags(c.buildCommand())
}

func run(cmd *cobra.Command, args []string) {
	slog.Debug("Executing exit command")
	os.Exit(0)
}
