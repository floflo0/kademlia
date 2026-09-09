package adapters

import (
	"errors"
	"kademlia/internal/core/ports"
	"sync"
)

type message struct {
	from    ports.Address
	payload []byte
}

type MockNetworkAdapter struct {
	mu        sync.RWMutex
	listeners map[ports.Address]chan message
	nextPort  int
}

type mockListenConnection struct {
	mu              sync.RWMutex
	network         *MockNetworkAdapter
	address         ports.Address
	messagesChannel chan message
	closed          bool
}

type mockDialConnection struct {
	mu                 sync.RWMutex
	network            *MockNetworkAdapter
	address            ports.Address
	destinationAddress ports.Address
	messagesChannel    chan message
	closed             bool
}

const listenersMessageChannelCapacity = 10

func NewMockNetworkAdapter() *MockNetworkAdapter {
	return &MockNetworkAdapter{
		listeners: make(map[ports.Address]chan message),
		nextPort:  10_000,
	}
}

func (n *MockNetworkAdapter) Listen(address ports.Address) (ports.ListenConnection, error) {
	n.mu.Lock()
	defer n.mu.Unlock()
	if _, exists := n.listeners[address]; exists {
		return nil, errors.New("address already in use")
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

func (n *MockNetworkAdapter) Dial(address ports.Address) (ports.DialConnection, error) {
	n.mu.Lock()
	defer n.mu.Unlock()
	assignedAddress := ports.Address{
		IP:   "127.0.0.1",
		Port: n.nextPort,
	}
	n.nextPort++
	if _, exists := n.listeners[assignedAddress]; exists {
		return nil, errors.New("address already in use")
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

func (n *MockNetworkAdapter) send(from ports.Address, to ports.Address, payload []byte) error {
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
			return errors.New("message queue is full")
		}
	}
	return errors.New("destination address not found")
}

func (c *mockListenConnection) SendTo(address ports.Address, payload []byte) error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.closed {
		return errors.New("use of closed network connection")
	}
	return c.network.send(c.address, address, payload)
}

func (c *mockListenConnection) Receive() ([]byte, *ports.Address, error) {
	c.mu.RLock()
	closed := c.closed
	c.mu.RUnlock()
	if closed {
		return nil, nil, errors.New("use of closed network connection")
	}
	message, ok := <-c.messagesChannel
	if !ok {
		return nil, nil, errors.New("connection closed")
	}
	return message.payload, &message.from, nil
}

func (c *mockListenConnection) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return errors.New("use of closed network connection")
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
		return errors.New("use of closed network connection")
	}
	return c.network.send(c.address, c.destinationAddress, payload)
}

func (c *mockDialConnection) Receive(timeoutMiliseconds uint32) ([]byte, error) {
	c.mu.RLock()
	closed := c.closed
	c.mu.RUnlock()
	if closed {
		return nil, errors.New("use of closed network connection")
	}
	message, ok := <-c.messagesChannel
	if !ok {
		return nil, errors.New("connection closed")
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
		return errors.New("use of closed network connection")
	}
	c.closed = true

	c.network.mu.Lock()
	defer c.network.mu.Unlock()
	delete(c.network.listeners, c.address)
	close(c.messagesChannel)
	return nil
}
