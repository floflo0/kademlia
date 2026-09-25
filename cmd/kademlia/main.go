package main

import (
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

		kademlia := kademlia.NewKademlia(
			kademlia.NewContactFromAddress(entities.Address{
				IP:   config.Host,
				Port: config.Port,
			}),
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

		return kademlia.Run(nil)
	})
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
