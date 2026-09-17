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

		kademlia0 := kademlia.NewKademlia(
			kademlia.NewContact(
				kademlia.NewKademliaID("0000000000000000000000000000000000000000000000000000000000000001"),
				entities.Address{
					IP:   "127.0.0.1",
					Port: 8000,
				},
			),
			network,
		)
		kademlia1 := kademlia.NewKademlia(
			kademlia.NewContact(
				kademlia.NewKademliaID("0000000000000000000000000000000000000000000000000000000000000050"),
				entities.Address{
					IP:   "127.0.0.1",
					Port: 8001,
				},
			),
			network,
		)
		kademlia2 := kademlia.NewKademlia(
			kademlia.NewContact(
				kademlia.NewKademliaID("0000000000000000000000000000000000000000000000000000000000000010"),
				entities.Address{
					IP:   "127.0.0.1",
					Port: 8002,
				},
			),
			network,
		)
		kademlia3 := kademlia.NewKademlia(
			kademlia.NewContact(
				kademlia.NewKademliaID("0000000000000000000000000000000000000000000000000000000000000020"),
				entities.Address{
					IP:   "127.0.0.1",
					Port: 8003,
				},
			),
			network,
		)
		kademlia0.RoutingTable.AddContact(kademlia.NewContact(
			kademlia.NewKademliaID("0000000000000000000000000000000000000000000000000000000000000050"),
			entities.Address{
				IP:   "127.0.0.1",
				Port: 8001,
			},
		))
		kademlia0.RoutingTable.AddContact(kademlia.NewContact(
			kademlia.NewKademliaID("0000000000000000000000000000000000000000000000000000000000000010"),
			entities.Address{
				IP:   "127.0.0.1",
				Port: 8002,
			},
		))
		kademlia1.RoutingTable.AddContact(kademlia.NewContact(
			kademlia.NewKademliaID("0000000000000000000000000000000000000000000000000000000000000001"),
			entities.Address{
				IP:   "127.0.0.1",
				Port: 8000,
			},
		))
		kademlia2.RoutingTable.AddContact(kademlia.NewContact(
			kademlia.NewKademliaID("0000000000000000000000000000000000000000000000000000000000000020"),
			entities.Address{
				IP:   "127.0.0.1",
				Port: 8003,
			},
		))

		target := kademlia.NewKademliaID("0000000000000000000000000000000000000000000000000000000000000016")
		go kademlia0.Run()
		go kademlia1.Run()
		go kademlia2.Run()
		go kademlia3.Run()

		candidates, _ := kademlia0.LookupContact(target)
		slog.Debug("Candidates", "candidates", candidates)

		shell := shell.NewShell(kademlia0)
		return shell.Run()
	})
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
