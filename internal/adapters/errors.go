package adapters

import (
	"errors"
)

var (
	ErrAddressAlreadyInUse = errors.New("address already in use")
	ErrDestinationNotFound = errors.New("destination address not found")
	ErrMessageQueueFull    = errors.New("message queue is full")
)
