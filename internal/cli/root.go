package cli

import (
	"kademlia/config"

	"github.com/spf13/cobra"
)

type Config struct {
	Host string
	Port int
}

func NewRootCommand(
	commandName string,
	start func(config Config) error,
) *cobra.Command {
	var appConfig Config
	rootCommand := &cobra.Command{
		Use:                   commandName + " [-h] [--host host] [-p port]",
		Short:                 "Kademlia node",
		Long:                  "Kademlia node",
		Example:               commandName + " --host 0.0.0.0 --port 8081",
		DisableFlagsInUseLine: true,
		Args:                  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
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
	return rootCommand
}
