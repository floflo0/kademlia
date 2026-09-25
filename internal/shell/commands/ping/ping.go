package ping

import (
	"fmt"
	"kademlia/config"
	"kademlia/internal/core/entities"
	"kademlia/internal/core/kademlia"
	"kademlia/internal/shell/commands"
	"log/slog"
	"strconv"

	"github.com/spf13/cobra"
)

type pingCommand struct {
	kademlia kademlia.Kademlia
}

func NewPingCommand(kademlia kademlia.Kademlia) commands.Command {
	return &pingCommand{
		kademlia: kademlia,
	}
}

func (c *pingCommand) buildCommand() *cobra.Command {
	command := &cobra.Command{
		Use:                   "ping [-h] adresss [port]",
		Short:                 "Ping a node",
		Args:                  cobra.RangeArgs(1, 2),
		Example:               "ping 127.0.0.1 8081",
		DisableFlagsInUseLine: true,
		RunE: func(command *cobra.Command, args []string) error {
			port := config.DefaultPort
			if len(args) == 2 {
				var err error
				port, err = strconv.Atoi(args[1])
				if err != nil {
					return fmt.Errorf("invalid port: %q: %v", args[1], err)
				}
			}
			address := entities.Address{
				IP:   args[0],
				Port: port,
			}
			elapsedTime, err := c.kademlia.Ping(address)
			if err != nil {
				slog.Error("Ping", "err", err)
				return nil
			}
			slog.Info(
				"Ping",
				"address", address,
				"elapsedTime", elapsedTime,
			)
			return nil
		},
	}
	return command
}

func (c *pingCommand) Execute(args []string) error {
	command := c.buildCommand()
	command.SetArgs(args)
	return command.Execute()
}

func (c *pingCommand) GetCompletions() []commands.Completion {
	return commands.GetCompletions(c.buildCommand())
}
