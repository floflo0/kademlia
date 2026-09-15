package adapters_test

import (
	"errors"
	"kademlia/internal/adapters"
	"kademlia/internal/core/entities"
	"testing"
	"time"
)

func TestNewMockNetworkAdapter(t *testing.T) {
	network := adapters.NewMockNetworkAdapter()
	if network == nil {
		t.Fatal("NewMockNetworkAdapter() returned nil")
	}
}

func TestMockNetworkAdapter_Listen_Success(t *testing.T) {
	network := adapters.NewMockNetworkAdapter()

	address := entities.Address{
		IP:   "127.0.0.1",
		Port: 8000,
	}
	connection, err := network.Listen(address)
	if err != nil {
		t.Fatalf("Listen(%v) returned unexpected error: %v", address, err)
	}
	if connection == nil {
		t.Fatal("Listen() returned nil connection")
	}
}

func TestMockNetworkAdapter_Listen_DifferentAddresses(t *testing.T) {
	network := adapters.NewMockNetworkAdapter()

	address1 := entities.Address{
		IP:   "127.0.0.1",
		Port: 8000,
	}
	_, err := network.Listen(address1)
	if err != nil {
		t.Fatalf("Listen(%v) returned unexpected error: %v", address1, err)
	}

	address2 := entities.Address{
		IP:   "127.0.0.1",
		Port: 8001,
	}
	_, err = network.Listen(address2)
	if err != nil {
		t.Fatalf("Listen(%v) returned unexpected error: %v", address2, err)
	}
}

func TestMockNetworkAdapter_Listen_ReuseAddress(t *testing.T) {
	network := adapters.NewMockNetworkAdapter()

	address := entities.Address{
		IP:   "127.0.0.1",
		Port: 8000,
	}
	connection, err := network.Listen(address)
	if err != nil {
		t.Fatalf("Listen(%v) returned unexpected error: %v", address, err)
	}

	err = connection.Close()
	if err != nil {
		t.Fatalf("Close() returned unexpected error: %v", err)
	}

	_, err = network.Listen(address)
	if err != nil {
		t.Fatalf("Listen(%v) returned unexpected error: %v", address, err)
	}
}

func TestMockNetworkAdapter_Listen_AddressAlreadyInUse(t *testing.T) {
	network := adapters.NewMockNetworkAdapter()

	address := entities.Address{
		IP:   "127.0.0.1",
		Port: 8000,
	}
	_, err := network.Listen(address)
	if err != nil {
		t.Fatalf("Listen(%v) returned unexpected error: %v", address, err)
	}

	_, err = network.Listen(address)
	if !errors.Is(err, adapters.ErrAddressAlreadyInUse) {
		t.Fatalf(
			"Listen(%v) err = %v; want %v",
			address,
			err,
			adapters.ErrAddressAlreadyInUse,
		)
	}
}

func TestMockNetworkAdapter_Listen_SendTo_Receive_Success(t *testing.T) {
	network := adapters.NewMockNetworkAdapter()

	address1 := entities.Address{IP: "127.0.0.1", Port: 8000}
	address2 := entities.Address{IP: "127.0.0.1", Port: 8001}

	connection1, err := network.Listen(address1)
	if err != nil {
		t.Fatalf("Listen(%v) returned unexpected error: %v", address1, err)
	}

	connection2, err := network.Listen(address2)
	if err != nil {
		t.Fatalf("Listen(%v) returned unexpected error: %v", address2, err)
	}

	payload := []byte("hello")

	err = connection1.SendTo(address2, payload)
	if err != nil {
		t.Fatalf(
			"SendTo(%v, %v) returned unexpected error: %v",
			address2,
			payload,
			err,
		)
	}

	received, from, err := connection2.Receive()
	if err != nil {
		t.Fatalf("Receive() returned unexpected error: %v", err)
	}

	if string(received) != string(payload) {
		t.Fatalf("Receive() payload = %q; want %q", received, payload)
	}

	if *from != address1 {
		t.Fatalf("Receive() sender = %v; want %v", *from, address1)
	}
}

func TestMockNetworkAdapter_Listen_SendTo_DestinationNotFound(t *testing.T) {
	network := adapters.NewMockNetworkAdapter()

	address1 := entities.Address{
		IP:   "127.0.0.1",
		Port: 8000,
	}
	connection, err := network.Listen(address1)
	if err != nil {
		t.Fatalf("Listen(%v) returned unexpected error: %v", address1, err)
	}

	address2 := entities.Address{
		IP:   "127.0.0.1",
		Port: 8001,
	}
	payload := []byte("hello")
	err = connection.SendTo(address2, payload)
	if !errors.Is(err, adapters.ErrDestinationNotFound) {
		t.Fatalf(
			"SendTo(%v, %v) err = %v; want %v",
			address2,
			payload,
			err,
			adapters.ErrDestinationNotFound,
		)
	}
}

func TestMockNetworkAdapter_Listen_SendTo_AfterClose(t *testing.T) {
	network := adapters.NewMockNetworkAdapter()

	address1 := entities.Address{
		IP:   "127.0.0.1",
		Port: 8000,
	}
	connection, err := network.Listen(address1)
	if err != nil {
		t.Fatalf("Listen(%v) returned unexpected error: %v", address1, err)
	}

	err = connection.Close()
	if err != nil {
		t.Fatalf("Close() returned unexpected error: %v", err)
	}

	address2 := entities.Address{
		IP:   "127.0.0.1",
		Port: 8001,
	}
	payload := []byte("hello")
	err = connection.SendTo(address2, payload)
	if !errors.Is(err, adapters.ErrClosedNetworkConnection) {
		t.Fatalf(
			"SendTo(%v, %v) err = %v; want %v",
			address2,
			payload,
			err,
			adapters.ErrClosedNetworkConnection,
		)
	}
}

func TestMockNetworkAdapter_Listen_SendTo_QueueFull(t *testing.T) {
	network := adapters.NewMockNetworkAdapter()

	address := entities.Address{
		IP:   "127.0.0.1",
		Port: 8000,
	}
	connection, err := network.Listen(address)
	if err != nil {
		t.Fatalf("Listen(%v) returned unexpected error: %v", address, err)
	}

	destinationAddress := entities.Address{
		IP:   "127.0.0.1",
		Port: 8001,
	}
	_, err = network.Listen(destinationAddress)
	if err != nil {
		t.Fatalf(
			"Listen(%v) returned unexpected error: %v",
			destinationAddress,
			err,
		)
	}

	payload := []byte("hello")
	for i := 0; i < 10; i++ {
		err = connection.SendTo(destinationAddress, payload)
		if err != nil {
			t.Fatalf(
				"SendTo(%v, %v) returned unexpected error: %v",
				destinationAddress,
				payload,
				err,
			)
		}
	}

	err = connection.SendTo(destinationAddress, payload)
	if !errors.Is(err, adapters.ErrMessageQueueFull) {
		t.Fatalf(
			"SendTo(%v, %v) err = %v; want %v",
			destinationAddress,
			payload,
			err,
			adapters.ErrMessageQueueFull,
		)
	}
}

func TestMockNetworkAdapter_Listen_Receive_AfterClose(t *testing.T) {
	network := adapters.NewMockNetworkAdapter()

	address := entities.Address{
		IP:   "127.0.0.1",
		Port: 8000,
	}
	connection, err := network.Listen(address)
	if err != nil {
		t.Fatalf("Listen(%v) returned unexpected error: %v", address, err)
	}

	err = connection.Close()
	if err != nil {
		t.Fatalf("Close() returned unexpected error: %v", err)
	}

	_, _, err = connection.Receive()
	if !errors.Is(err, adapters.ErrClosedNetworkConnection) {
		t.Fatalf(
			"Receive() err = %v; want %v",
			err,
			adapters.ErrClosedNetworkConnection,
		)
	}
}

func TestMockNetworkAdapter_Listen_Receive_ConnectionClosed(t *testing.T) {
	network := adapters.NewMockNetworkAdapter()

	address := entities.Address{
		IP:   "127.0.0.1",
		Port: 8000,
	}
	connection, err := network.Listen(address)
	if err != nil {
		t.Fatalf("Listen(%v) returned unexpected error: %v", address, err)
	}

	receiveStarted := make(chan struct{})
	receiveErr := make(chan error)

	go func() {
		close(receiveStarted)
		_, _, err := connection.Receive()
		receiveErr <- err
	}()

	<-receiveStarted
	// Give time to Receive to start waiting for the channel.
	time.Sleep(time.Millisecond)
	err = connection.Close()
	if err != nil {
		t.Fatalf("Close() returned unexpected error: %v", err)
	}

	select {
	case err = <-receiveErr:
		if !errors.Is(err, adapters.ErrConnectionClosed) {
			t.Fatalf(
				"Receive() err = %v; want %v",
				err,
				adapters.ErrConnectionClosed,
			)
		}
	case <-time.After(time.Second):
		t.Fatalf("Receive() did not return after connection was closed")
	}
}

func TestMockNetworkAdapter_Listen_Close_Sucess(t *testing.T) {
	network := adapters.NewMockNetworkAdapter()

	address := entities.Address{
		IP:   "127.0.0.1",
		Port: 8000,
	}
	connection, err := network.Listen(address)
	if err != nil {
		t.Fatalf("Listen(%v) returned unexpected error: %v", address, err)
	}

	err = connection.Close()
	if err != nil {
		t.Fatalf("Close() returned unexpected error: %v", err)
	}
}

func TestMockNetworkAdapter_Listen_Close_AlreadyClosed(t *testing.T) {
	network := adapters.NewMockNetworkAdapter()

	address := entities.Address{
		IP:   "127.0.0.1",
		Port: 8000,
	}
	connection, err := network.Listen(address)
	if err != nil {
		t.Fatalf("Listen(%v) returned unexpected error: %v", address, err)
	}

	err = connection.Close()
	if err != nil {
		t.Fatalf("Close() returned unexpected error: %v", err)
	}

	err = connection.Close()
	if !errors.Is(err, adapters.ErrClosedNetworkConnection) {
		t.Fatalf(
			"Close()err = %v; want %v",
			err,
			adapters.ErrClosedNetworkConnection,
		)
	}
}

func TestMockNetworkAdapter_Dial_Success(t *testing.T) {
	network := adapters.NewMockNetworkAdapter()

	address := entities.Address{
		IP:   "127.0.0.1",
		Port: 8000,
	}
	connection, err := network.Dial(address)
	if err != nil {
		t.Fatalf("Dial(%v) returned unexpected error: %v", address, err)
	}
	if connection == nil {
		t.Fatal("Dial() returned nil connection")
	}
}

func TestMockNetworkAdapter_Dial_SameAddress(t *testing.T) {
	network := adapters.NewMockNetworkAdapter()
	address := entities.Address{
		IP:   "127.0.0.1",
		Port: 8000,
	}
	_, err := network.Dial(address)
	if err != nil {
		t.Fatalf("Dial(%v) returned unexpected error: %v", address, err)
	}

	_, err = network.Dial(address)
	if err != nil {
		t.Fatalf("Dial(%v) returned unexpected error: %v", address, err)
	}
}

func TestMockNetworkAdapter_Dial_AddressAlreadyInUse(t *testing.T) {
	network := adapters.NewMockNetworkAdapter()

	address1 := entities.Address{
		IP:   "127.0.0.1",
		Port: 10_000,
	}
	_, err := network.Listen(address1)
	if err != nil {
		t.Fatalf("Listen(%v) returned unexpected error: %v", address1, err)
	}

	address2 := entities.Address{
		IP:   "127.0.0.1",
		Port: 8000,
	}
	_, err = network.Dial(address2)
	if !errors.Is(err, adapters.ErrAddressAlreadyInUse) {
		t.Fatalf(
			"Dial(%v) err = %v; want %v",
			address2,
			err,
			adapters.ErrAddressAlreadyInUse,
		)
	}
}

func TestMockNetworkAdapter_Dial_Send_Receive_Success(t *testing.T) {
	network := adapters.NewMockNetworkAdapter()

	address := entities.Address{
		IP:   "127.0.0.1",
		Port: 8000,
	}
	listenConnection, err := network.Listen(address)
	if err != nil {
		t.Fatalf("Listen(%v) returned unexpected error: %v", address, err)
	}

	dialConnection, err := network.Dial(address)
	if err != nil {
		t.Fatalf("Dial(%v) returned unexpected error: %v", address, err)
	}

	payload := []byte("hello")
	responsePlayload := []byte("response")

	go func() {
		received, address, err := listenConnection.Receive()
		if err != nil {
			t.Errorf("Receive() returned unexpected error: %v", err)
			return
		}
		if string(received) != string(payload) {
			t.Errorf("Receive() payload = %q; want %q", received, payload)
			return
		}

		err = listenConnection.SendTo(*address, responsePlayload)
		if err != nil {
			t.Errorf(
				"SendTo(%v, %q) returned unexpected error: %v",
				*address,
				string(responsePlayload),
				err,
			)
			return
		}
	}()

	err = dialConnection.Send(payload)
	if err != nil {
		t.Fatalf("Send(%q) returned unexpected error: %v", string(payload), err)
	}
	timeout := uint32(0)
	received, err := dialConnection.Receive(timeout)
	if err != nil {
		t.Fatalf("Receive(%v) returned unexpected error: %v", timeout, err)
	}
	if string(received) != string(responsePlayload) {
		t.Fatalf(
			"Receive(%v) payload = %q; want %q",
			timeout,
			received,
			responsePlayload,
		)
	}
}

func TestMockNetworkAdapter_Dial_Send_AfterClose(t *testing.T) {
	network := adapters.NewMockNetworkAdapter()

	address := entities.Address{
		IP:   "127.0.0.1",
		Port: 8000,
	}
	connection, err := network.Dial(address)
	if err != nil {
		t.Fatalf("Dial(%v) returned unexpected error: %v", address, err)
	}

	err = connection.Close()
	if err != nil {
		t.Fatalf("Close() returned unexpected error: %v", err)
	}

	payload := []byte("hello")
	err = connection.Send(payload)
	if !errors.Is(err, adapters.ErrClosedNetworkConnection) {
		t.Fatalf(
			"Send(%v) err = %v; want %v",
			payload,
			err,
			adapters.ErrClosedNetworkConnection,
		)
	}
}

func TestMockNetworkAdapter_Dial_Receive_AfterClose(t *testing.T) {
	network := adapters.NewMockNetworkAdapter()

	address := entities.Address{
		IP:   "127.0.0.1",
		Port: 8000,
	}
	connection, err := network.Dial(address)
	if err != nil {
		t.Fatalf("Dial(%v) returned unexpected error: %v", address, err)
	}

	err = connection.Close()
	if err != nil {
		t.Fatalf("Close() returned unexpected error: %v", err)
	}

	timeout := uint32(0)
	_, err = connection.Receive(timeout)
	if !errors.Is(err, adapters.ErrClosedNetworkConnection) {
		t.Fatalf(
			"Receive(%v) err = %v; want %v",
			timeout,
			err,
			adapters.ErrClosedNetworkConnection,
		)
	}
}

func TestMockNetworkAdapter_Dial_Receive_ConnectionClosed(t *testing.T) {
	network := adapters.NewMockNetworkAdapter()

	address := entities.Address{
		IP:   "127.0.0.1",
		Port: 8000,
	}
	connection, err := network.Dial(address)
	if err != nil {
		t.Fatalf("Dial(%v) returned unexpected error: %v", address, err)
	}

	timeout := uint32(0)
	receiveStarted := make(chan struct{})
	receiveErr := make(chan error)

	go func() {
		close(receiveStarted)
		_, err := connection.Receive(timeout)
		receiveErr <- err
	}()

	<-receiveStarted
	// Give time to Receive to start waiting for the channel.
	time.Sleep(time.Millisecond)
	err = connection.Close()
	if err != nil {
		t.Fatalf("Close() returned unexpected error: %v", err)
	}

	select {
	case err = <-receiveErr:
		if !errors.Is(err, adapters.ErrConnectionClosed) {
			t.Fatalf(
				"Receive(%v) err = %v; want %v",
				timeout,
				err,
				adapters.ErrConnectionClosed,
			)
		}

	case <-time.After(time.Second):
		t.Fatalf(
			"Receive(%v) did not return after connection was closed",
			timeout,
		)
	}
}

func TestMockNetworkAdapter_Dial_Close_Sucess(t *testing.T) {
	network := adapters.NewMockNetworkAdapter()

	address := entities.Address{
		IP:   "127.0.0.1",
		Port: 8000,
	}
	connection, err := network.Dial(address)
	if err != nil {
		t.Fatalf("Dial(%v) returned unexpected error: %v", address, err)
	}

	err = connection.Close()
	if err != nil {
		t.Fatalf("Close() returned unexpected error: %v", err)
	}
}

func TestMockNetworkAdapter_Dial_Close_AlreadyClosed(t *testing.T) {
	network := adapters.NewMockNetworkAdapter()

	address := entities.Address{
		IP:   "127.0.0.1",
		Port: 8000,
	}
	connection, err := network.Dial(address)
	if err != nil {
		t.Fatalf("Dial(%v) returned unexpected error: %v", address, err)
	}

	err = connection.Close()
	if err != nil {
		t.Fatalf("Close() returned unexpected error: %v", err)
	}

	err = connection.Close()
	if !errors.Is(err, adapters.ErrClosedNetworkConnection) {
		t.Fatalf(
			"Close() err = %v; want %v",
			err,
			adapters.ErrClosedNetworkConnection,
		)
	}
}
