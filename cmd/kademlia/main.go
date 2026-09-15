package main

import (
	"kademlia/internal/cli"
	"kademlia/internal/logging"
	"kademlia/internal/shell"
	"os"
)

func main() {
	logging.InitLogger()

	rootCmd := cli.NewRootCmd(os.Args[0], shell.NewShell())
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
