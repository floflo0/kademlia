package main

import (
	"kademlia/internal/adapters"
	"kademlia/internal/cli"
	"kademlia/internal/core/entities"
	"kademlia/internal/core/kademlia"
	"kademlia/internal/logging"
	"kademlia/internal/shell"
	"os"
)

func main() {
	logging.InitLogger(os.Stdout)

	rootCmd := cli.NewRootCommand(os.Args[0], func(config cli.Config) error {
		network := adapters.NewUDPNetworkAdapter()

		kademlia := kademlia.NewKademlia(
			kademlia.NewContact(
				kademlia.NewKademliaID("0000000000000000000000000000000000000000000000000000000000000001"),
				entities.Address{
					IP:   config.Host,
					Port: config.Port,
				},
			),
			network,
		)
		go kademlia.Run()

		shell := shell.NewShell(kademlia)
		return shell.Run()
	})
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
