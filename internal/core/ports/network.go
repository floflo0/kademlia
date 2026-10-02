package ports

import (
	"errors"
	"kademlia/internal/core/entities"
)

type Network interface {
	Listen(address entities.Address) (ListenConnection, error)
	Dial(address entities.Address) (DialConnection, error)
	GetIP() (string, error)
}

type ListenConnection interface {
	SendTo(address entities.Address, payload []byte) error
	Receive() ([]byte, *entities.Address, error)
	Close() error
}

type DialConnection interface {
	Send(payload []byte) error
	Receive(timeoutMiliseconds uint32) ([]byte, error)
	Close() error
}

var (
	ErrClosedNetworkConnection = errors.New("use of closed network connection")
	ErrTimeout                 = errors.New("network timeout")
)
