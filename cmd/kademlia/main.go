package main

import (
	"kademlia/internal/logging"
	"log/slog"
)

func main() {
	logging.InitLogger()

	slog.Info("Hello, World!")
}
