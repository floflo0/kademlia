package commands

import (
	"log/slog"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
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
	command := c.buildCommand()
	command.InitDefaultHelpFlag()
	flags := make([]string, 0)
	command.Flags().VisitAll(func(flag *pflag.Flag) {
		flags = append(flags, "--"+flag.Name)
		if flag.Shorthand != "" {
			flags = append(flags, "-"+flag.Shorthand)
		}
	})
	return flags
}

func run(cmd *cobra.Command, args []string) {
	slog.Debug("Executing exit command")
	os.Exit(0)
}
