package cli

import (
	"fmt"
	"kademlia/config"
	"kademlia/internal/adapters"
	"kademlia/internal/core/kademlia"
	"kademlia/internal/core/ports"
	"log/slog"

	"github.com/spf13/cobra"
)

func NewRootCmd(commandName string) *cobra.Command {
	rootCmd := &cobra.Command{
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

			//Put code here
			network := adapters.NewUDPNetworkAdapter()
			kad := kademlia.NewKademlia(kademlia.NewContact(kademlia.NewKademliaID("0000000000000000000000000000000000000000000000000000000000000001"), ports.Address{IP: "127.0.0.1", Port: 8080}), network)
			kad.RoutingTable.AddContact(kademlia.NewContact(kademlia.NewKademliaID("0000000000000000000000000000000000000000000000000000000000000002"), ports.Address{IP: "127.0.0.1", Port: 8081}))
			r := kad.LookupContact(kademlia.NewKademliaID("0000000000000000000000000000000000000000000000000000000000000003"))
			slog.Debug("", "r", r)
			return nil
		},
	}
	rootCmd.Flags().String("host", config.DefaultHost, "the host")
	rootCmd.Flags().IntP("port", "p", config.DefaultPort, "the port")
	return rootCmd
}
