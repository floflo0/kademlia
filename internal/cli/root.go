package cli

import (
	"kademlia/config"
	"kademlia/internal/core/entities"

	"github.com/spf13/cobra"
)

type Config struct {
	Host         string
	Port         int
	KnownContact *entities.Address
}

func NewRootCommand(
	commandName string,
	start func(config Config) error,
) *cobra.Command {
	var appConfig Config
	var knownContact string
	rootCommand := &cobra.Command{
		Use:                   commandName + " [-h] [--host host] [-p port]",
		Short:                 "Kademlia node",
		Long:                  "Kademlia node",
		Example:               commandName + " --host 0.0.0.0 --port 8081",
		DisableFlagsInUseLine: true,
		Args:                  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if knownContact != "" {
				knownContactAddress, err := entities.NewAddressFromString(
					knownContact,
				)
				if err != nil {
					return err
				}
				appConfig.KnownContact = knownContactAddress
			}
			return start(appConfig)
		},
	}
	rootCommand.Flags().StringVar(
		&appConfig.Host,
		"host",
		config.DefaultHost,
		"the host",
	)
	rootCommand.Flags().IntVarP(
		&appConfig.Port,
		"port",
		"p",
		config.DefaultPort,
		"the port",
	)
	rootCommand.Flags().StringVar(
		&knownContact,
		"known-contact",
		"",
		"the address to a known contact use to join the network (IP:PORT)",
	)
	return rootCommand
}
