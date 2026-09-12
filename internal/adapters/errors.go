package adapters

import "errors"

var (
	ErrAddressAlreadyInUse     = errors.New("address already in use")
	ErrClosedNetworkConnection = errors.New("use of closed network connection")
	ErrConnectionClosed        = errors.New("connection closed")
	ErrDestinationNotFound     = errors.New("destination address not found")
	ErrMessageQueueFull        = errors.New("message queue is full")
)
