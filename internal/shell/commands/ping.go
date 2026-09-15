package commands

import (
	"kademlia/internal/core/entities"
	. "kademlia/internal/core/kademlia"
	"log/slog"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type pingCommand struct{
	kademlia Kademlia
}

func NewPingCommand(kademlia Kademlia) Command {
	return &pingCommand{
		kademlia: kademlia,
	}
}

func (c *pingCommand) buildCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "ping [-h]",
		Short:                 "TODO",
		Args:                  cobra.NoArgs,
		DisableFlagsInUseLine: true,
		Run: func(cmd *cobra.Command, args []string) {
			elapsedTime, err := c.kademlia.Ping(entities.Address{
				IP: "127.0.0.1",
				Port: 8080,
			})
			slog.Debug("ping", "elapsedTime", elapsedTime, "err", err)
		},
	}
	return cmd
}

func (c *pingCommand) Execute(args []string) error {
	command := c.buildCommand()
	command.SetArgs(args)
	return command.Execute()
}

func (c *pingCommand) GetFlags() []string {
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
