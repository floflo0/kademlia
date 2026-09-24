package commands

import (
	"fmt"
	"kademlia/internal/core/kademlia"
	"log/slog"

	"github.com/spf13/cobra"
)

type showCommand struct {
	kademlia kademlia.Kademlia
}

func NewShowCommand(kademlia kademlia.Kademlia) Command {
	return &showCommand{
		kademlia: kademlia,
	}
}

func (c *showCommand) buildCommand() *cobra.Command {
	command := &cobra.Command{
		Use:                   "show [-h] rt|ds",
		Short:                 "Show debug information",
		Args:                  cobra.NoArgs,
		Example:               "show rt",
		DisableFlagsInUseLine: true,
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
	}
	command.AddCommand(
		showRoutingTableCommand(c.kademlia),
		showDataStoreCommand(c.kademlia),
	)
	return command
}

func (c *showCommand) Execute(args []string) error {
	command := c.buildCommand()
	command.SetArgs(args)
	return command.Execute()
}

func (c *showCommand) GetCompletions() []Completion {
	return getCompletions(c.buildCommand())
}

func showRoutingTableCommand(kademlia kademlia.Kademlia) *cobra.Command {
	command := &cobra.Command{
		Use:   "rt [-h]",
		Short: "Show the routing table",
		Args:  cobra.NoArgs,
		RunE: func(command *cobra.Command, args []string) error {
			slog.Debug("Running show rt command")

			routingTableEmpty := true
			for i, bucket := range kademlia.GetBuckets() {
				if bucket.Len() == 0 {
					continue
				}
				routingTableEmpty = false
				fmt.Printf("Bucket %d:\n", i)
				for j, contact := range bucket.GetContacts() {
					fmt.Printf(
						"    - %2d %s %s\n",
						j,
						contact.ID.ShortString(),
						contact.Address.String(),
					)
				}
			}
			if routingTableEmpty {
				fmt.Println("Routing table is empty.")
				return nil
			}
			return nil
		},
	}
	return command
}

func showDataStoreCommand(kademlia kademlia.Kademlia) *cobra.Command {
	command := &cobra.Command{
		Use:   "ds [-h]",
		Short: "Show the data store keys",
		Args:  cobra.NoArgs,
		RunE: func(command *cobra.Command, args []string) error {
			slog.Debug("Running show ds command")
			for _, key := range kademlia.GetStoredKeys() {
				fmt.Println(key)
			}
			return nil
		},
	}
	return command
}
