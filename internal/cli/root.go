package cli

import (
	"fmt"
	"kademlia/config"
	"kademlia/internal/shell"
	"log/slog"

	"github.com/spf13/cobra"
)

func NewRootCmd(commandName string) *cobra.Command {
	rootCommand := &cobra.Command{
		Use:                   commandName + " [-h] [--host host] [-p port]",
		Short:                 "Kademlia node",
		Long:                  "Kademlia node",
		Example:               commandName + " --host 0.0.0.0 --port 8081",
		DisableFlagsInUseLine: true,
		Args:                  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			host, err := cmd.Flags().GetString("host")
			if err != nil {
				return fmt.Errorf("failed to get host parameter: %w", err)
			}

			port, err := cmd.Flags().GetInt("port")
			if err != nil {
				return fmt.Errorf("failed to get port parameter: %w", err)
			}

			slog.Debug("Parsed command line", "host", host, "port", port)
			return shell.NewShell().Run()
		},
	}
	rootCommand.Flags().String("host", config.DefaultHost, "the host")
	rootCommand.Flags().IntP("port", "p", config.DefaultPort, "the port")
	return rootCommand
}
