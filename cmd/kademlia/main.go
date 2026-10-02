package main

import (
	"fmt"
	"kademlia/internal/adapters"
	"kademlia/internal/cli"
	"kademlia/internal/core/entities"
	"kademlia/internal/core/kademlia"
	"kademlia/internal/logging"
	"kademlia/internal/shell"
	"log/slog"
	"os"
)

func main() {
	logging.InitLogger(os.Stdout)

	rootCmd := cli.NewRootCommand(os.Args[0], func(config cli.Config) error {
		network := adapters.NewUDPNetworkAdapter()

		host := config.Host
		if host == "" {
			var err error
			host, err = network.GetIP()
			if err != nil {
				return fmt.Errorf("failed to get IP: %w", err)
			}
		}
		kademlia := kademlia.NewKademlia(
			entities.Address{
				IP:   host,
				Port: config.Port,
			},
			network,
		)

		go func() {
			shell := shell.NewShell(kademlia)
			err := shell.Run()
			if err != nil {
				slog.Error("Failed to run the shell", "err", err)
				return
			}
		}()

		return kademlia.Run(config.KnownContact)
	})
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
