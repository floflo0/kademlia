package main

import (
	"kademlia/internal/cli"
	"kademlia/internal/logging"
	"os"
)

func main() {
	logging.InitLogger()

	rootCmd := cli.NewRootCmd(os.Args[0])
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
