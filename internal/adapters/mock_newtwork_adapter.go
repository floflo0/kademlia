package adapters

import (
	"kademlia/internal/core/entities"
	"kademlia/internal/core/ports"
	"sync"
)

type message struct {
	from    entities.Address
	payload []byte
}

type MockNetworkAdapter struct {
	mu        sync.RWMutex
	listeners map[entities.Address]chan message
	nextPort  int
}

type mockListenConnection struct {
	mu              sync.RWMutex
	network         *MockNetworkAdapter
	address         entities.Address
	messagesChannel chan message
	closed          bool
}

type mockDialConnection struct {
	mu                 sync.RWMutex
	network            *MockNetworkAdapter
	address            entities.Address
	destinationAddress entities.Address
	messagesChannel    chan message
	closed             bool
}

const listenersMessageChannelCapacity = 10

func NewMockNetworkAdapter() *MockNetworkAdapter {
	return &MockNetworkAdapter{
		listeners: make(map[entities.Address]chan message),
		nextPort:  10_000,
	}
}

func (n *MockNetworkAdapter) Listen(address entities.Address) (ports.ListenConnection, error) {
	n.mu.Lock()
	defer n.mu.Unlock()
	if _, exists := n.listeners[address]; exists {
		return nil, ErrAddressAlreadyInUse
	}
	messagesChannel := make(chan message, listenersMessageChannelCapacity)
	n.listeners[address] = messagesChannel
	connection := &mockListenConnection{
		network:         n,
		address:         address,
		messagesChannel: messagesChannel,
		closed:          false,
	}
	return connection, nil
}

func (n *MockNetworkAdapter) Dial(address entities.Address) (ports.DialConnection, error) {
	n.mu.Lock()
	defer n.mu.Unlock()
	assignedAddress := entities.Address{
		IP:   "127.0.0.1",
		Port: n.nextPort,
	}
	n.nextPort++
	if _, exists := n.listeners[assignedAddress]; exists {
		return nil, ErrAddressAlreadyInUse
	}
	messagesChannel := make(chan message, listenersMessageChannelCapacity)
	n.listeners[assignedAddress] = messagesChannel
	connection := &mockDialConnection{
		network:            n,
		address:            assignedAddress,
		destinationAddress: address,
		messagesChannel:    messagesChannel,
		closed:             false,
	}
	return connection, nil
}

func (n *MockNetworkAdapter) send(from entities.Address, to entities.Address, payload []byte) error {
	n.mu.RLock()
	defer n.mu.RUnlock()
	if listener, exists := n.listeners[to]; exists {
		message := message{
			from:    from,
			payload: payload,
		}
		select {
		case listener <- message:
			return nil
		default:
			return ErrMessageQueueFull
		}
	}
	return ErrDestinationNotFound
}

func (c *mockListenConnection) GetIP() (string, error) {
	return c.address.IP, nil
}

func (c *mockListenConnection) SendTo(address entities.Address, payload []byte) error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.closed {
		return ErrClosedNetworkConnection
	}
	return c.network.send(c.address, address, payload)
}

func (c *mockListenConnection) Receive() ([]byte, *entities.Address, error) {
	c.mu.RLock()
	closed := c.closed
	c.mu.RUnlock()
	if closed {
		return nil, nil, ErrClosedNetworkConnection
	}
	message, ok := <-c.messagesChannel
	if !ok {
		return nil, nil, ErrConnectionClosed
	}
	return message.payload, &message.from, nil
}

func (c *mockListenConnection) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return ErrClosedNetworkConnection
	}
	c.closed = true

	c.network.mu.Lock()
	defer c.network.mu.Unlock()
	delete(c.network.listeners, c.address)
	close(c.messagesChannel)
	return nil
}

func (c *mockDialConnection) Send(payload []byte) error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.closed {
		return ErrClosedNetworkConnection
	}
	return c.network.send(c.address, c.destinationAddress, payload)
}

func (c *mockDialConnection) Receive(timeoutMiliseconds uint32) ([]byte, error) {
	c.mu.RLock()
	closed := c.closed
	c.mu.RUnlock()
	if closed {
		return nil, ErrClosedNetworkConnection
	}
	message, ok := <-c.messagesChannel
	if !ok {
		return nil, ErrConnectionClosed
	}
	if message.from != c.destinationAddress {
		panic("invalid source address")
	}
	return message.payload, nil
}

func (c *mockDialConnection) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return ErrClosedNetworkConnection
	}
	c.closed = true

	c.network.mu.Lock()
	defer c.network.mu.Unlock()
	delete(c.network.listeners, c.address)
	close(c.messagesChannel)
	return nil
}
