package adapters_test

import (
	"errors"
	"fmt"
	"kademlia/internal/adapters"
	"kademlia/internal/core/entities"
	"net"
	"testing"
	"time"
)

func TestNewUDPNetworkAdapter(t *testing.T) {
	network := adapters.NewUDPNetworkAdapter()
	if network == nil {
		t.Fatal("NewUDPNetworkAdapter() returned nil")
	}
}

func TestUDPNetworkAdapter_Listen_Success(t *testing.T) {
	network := adapters.NewUDPNetworkAdapter()

	address := entities.Address{IP: "127.0.0.1", Port: 0}
	connection, err := network.Listen(address)
	if err != nil {
		t.Fatalf("Listen(%v) returned unexpected error: %v", address, err)
	}
	if connection == nil {
		t.Fatalf("Listen(%v) returned nil connection", address)
	}
	defer connection.Close()
}

func TestUDPNetworkAdapter_Listen_InvalidIP(t *testing.T) {
	network := adapters.NewUDPNetworkAdapter()

	address := entities.Address{IP: "not-an-ip", Port: 20001}
	_, err := network.Listen(address)
	if err == nil {
		t.Fatalf("Listen(%v) err = nil; want an error", address)
	}
}

func TestUDPNetworkAdapter_Listen_DifferentAddresses(t *testing.T) {
	network := adapters.NewUDPNetworkAdapter()

	address1 := entities.Address{IP: "127.0.0.1", Port: 20002}
	connection1, err := network.Listen(address1)
	if err != nil {
		t.Fatalf("Listen(%v) returned unexpected error: %v", address1, err)
	}
	defer connection1.Close()

	address2 := entities.Address{IP: "127.0.0.1", Port: 20003}
	connection2, err := network.Listen(address2)
	if err != nil {
		t.Fatalf("Listen(%v) returned unexpected error: %v", address2, err)
	}
	defer connection2.Close()
}

func TestUDPNetworkAdapter_Listen_ReuseAddressAfterClose(t *testing.T) {
	network := adapters.NewUDPNetworkAdapter()

	address := entities.Address{IP: "127.0.0.1", Port: 20004}
	connection, err := network.Listen(address)
	if err != nil {
		t.Fatalf("Listen(%v) returned unexpected error: %v", address, err)
	}

	if err = connection.Close(); err != nil {
		t.Fatalf("Close() returned unexpected error: %v", err)
	}

	connection, err = network.Listen(address)
	if err != nil {
		t.Fatalf("Listen(%v) returned unexpected error: %v", address, err)
	}
	defer connection.Close()
}

func TestUDPNetworkAdapter_Listen_AddressAlreadyInUse(t *testing.T) {
	network := adapters.NewUDPNetworkAdapter()

	address := entities.Address{IP: "127.0.0.1", Port: 20005}
	connection, err := network.Listen(address)
	if err != nil {
		t.Fatalf("Listen(%v) returned unexpected error: %v", address, err)
	}
	defer connection.Close()

	_, err = network.Listen(address)
	if err == nil {
		t.Fatalf("Listen(%v) err = nil; want an error", address)
	}
}

func TestUDPNetworkAdapter_Listen_GetIP_Loopback(t *testing.T) {
	network := adapters.NewUDPNetworkAdapter()

	address := entities.Address{IP: "127.0.0.1", Port: 20006}
	connection, err := network.Listen(address)
	if err != nil {
		t.Fatalf("Listen(%v) returned unexpected error: %v", address, err)
	}
	defer connection.Close()

	ip, err := connection.GetIP()
	if err != nil {
		t.Fatalf("GetIP() returned unexpected error: %v", err)
	}
	expectedIP := "127.0.0.1"
	if ip != expectedIP {
		t.Fatalf("GetIP() ip = %v; want %v", ip, expectedIP)
	}
}

func TestUDPNetworkAdapter_Listen_GetIP_Unspecified(t *testing.T) {
	network := adapters.NewUDPNetworkAdapter()

	address := entities.Address{IP: "0.0.0.0", Port: 20007}
	connection, err := network.Listen(address)
	if err != nil {
		t.Fatalf("Listen(%v) returned unexpected error: %v", address, err)
	}
	defer connection.Close()

	ip, err := connection.GetIP()
	if err != nil {
		t.Fatalf("GetIP() returned unexpected error: %v", err)
	}
	if ip == "0.0.0.0" {
		t.Fatal("GetIP() returned an unspecified IP address")
	}
}

func TestUDPNetworkAdapter_Listen_SendTo_Receive_Success(t *testing.T) {
	network := adapters.NewUDPNetworkAdapter()

	address1 := entities.Address{IP: "127.0.0.1", Port: 20008}
	address2 := entities.Address{IP: "127.0.0.1", Port: 20009}

	connection1, err := network.Listen(address1)
	if err != nil {
		t.Fatalf("Listen(%v) returned unexpected error: %v", address1, err)
	}
	defer connection1.Close()

	connection2, err := network.Listen(address2)
	if err != nil {
		t.Fatalf("Listen(%v) returned unexpected error: %v", address2, err)
	}
	defer connection2.Close()

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

func TestUDPNetworkAdapter_Listen_SendTo_InvalidAddress(t *testing.T) {
	network := adapters.NewUDPNetworkAdapter()

	address := entities.Address{IP: "127.0.0.1", Port: 20010}
	connection, err := network.Listen(address)
	if err != nil {
		t.Fatalf("Listen(%v) returned unexpected error: %v", address, err)
	}
	defer connection.Close()

	destination := entities.Address{IP: "not-an-ip", Port: 20011}
	payload := []byte("hello")
	err = connection.SendTo(destination, payload)
	if err == nil {
		t.Fatalf(
			"SendTo(%v, %v) err = nil; want an error",
			destination,
			payload,
		)
	}
}

func TestUDPNetworkAdapter_Listen_SendTo_AfterClose(t *testing.T) {
	network := adapters.NewUDPNetworkAdapter()

	address1 := entities.Address{IP: "127.0.0.1", Port: 20012}
	connection, err := network.Listen(address1)
	if err != nil {
		t.Fatalf("Listen(%v) returned unexpected error: %v", address1, err)
	}

	if err = connection.Close(); err != nil {
		t.Fatalf("Close() returned unexpected error: %v", err)
	}

	address2 := entities.Address{IP: "127.0.0.1", Port: 20013}
	payload := []byte("hello")
	err = connection.SendTo(address2, payload)
	expectedError := net.ErrClosed
	if !errors.Is(err, expectedError) {
		t.Fatalf(
			"SendTo(%v, %v) err = %v; want %v",
			address2,
			payload,
			err,
			expectedError,
		)
	}
}

func TestUDPNetworkAdapter_Listen_Receive_AfterClose(t *testing.T) {
	network := adapters.NewUDPNetworkAdapter()

	address := entities.Address{IP: "127.0.0.1", Port: 20014}
	connection, err := network.Listen(address)
	if err != nil {
		t.Fatalf("Listen(%v) returned unexpected error: %v", address, err)
	}

	if err = connection.Close(); err != nil {
		t.Fatalf("Close() returned unexpected error: %v", err)
	}

	_, _, err = connection.Receive()
	expectedError := net.ErrClosed
	if !errors.Is(err, expectedError) {
		t.Fatalf("Receive() err = %v; want %v", err, expectedError)
	}
}

func TestUDPNetworkAdapter_Listen_Receive_ConnectionClosed(t *testing.T) {
	network := adapters.NewUDPNetworkAdapter()

	address := entities.Address{IP: "127.0.0.1", Port: 20015}
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
	// Give time for Receive to start blocking.
	time.Sleep(10 * time.Millisecond)
	if err = connection.Close(); err != nil {
		t.Fatalf("Close() returned unexpected error: %v", err)
	}

	select {
	case err = <-receiveErr:
		expectedError := net.ErrClosed
		if !errors.Is(err, expectedError) {
			t.Fatalf("Receive() err = %v; want %v", err, expectedError)
		}
	case <-time.After(time.Second):
		t.Fatal("Receive() did not return after connection was closed")
	}
}

func TestUDPNetworkAdapter_Listen_Close_Success(t *testing.T) {
	network := adapters.NewUDPNetworkAdapter()

	address := entities.Address{IP: "127.0.0.1", Port: 20016}
	connection, err := network.Listen(address)
	if err != nil {
		t.Fatalf("Listen(%v) returned unexpected error: %v", address, err)
	}

	if err = connection.Close(); err != nil {
		t.Fatalf("Close() returned unexpected error: %v", err)
	}
}

func TestUDPNetworkAdapter_Listen_Close_AlreadyClosed(t *testing.T) {
	network := adapters.NewUDPNetworkAdapter()

	address := entities.Address{IP: "127.0.0.1", Port: 20017}
	connection, err := network.Listen(address)
	if err != nil {
		t.Fatalf("Listen(%v) returned unexpected error: %v", address, err)
	}

	if err = connection.Close(); err != nil {
		t.Fatalf("Close() returned unexpected error: %v", err)
	}

	err = connection.Close()
	expectedError := net.ErrClosed
	if !errors.Is(err, expectedError) {
		t.Fatalf("Close() err = %v; want %v", err, expectedError)
	}
}

func TestUDPNetworkAdapter_Dial_Success(t *testing.T) {
	network := adapters.NewUDPNetworkAdapter()

	address := entities.Address{IP: "127.0.0.1", Port: 20018}
	connection, err := network.Dial(address)
	if err != nil {
		t.Fatalf("Dial(%v) returned unexpected error: %v", address, err)
	}
	if connection == nil {
		t.Fatalf("Dial(%v) returned nil connection", address)
	}
	defer connection.Close()
}

func TestUDPNetworkAdapter_Dial_InvalidIP(t *testing.T) {
	network := adapters.NewUDPNetworkAdapter()

	address := entities.Address{IP: "not-an-ip", Port: 20019}
	_, err := network.Dial(address)
	if err == nil {
		t.Fatalf("Dial(%v) err = nil; want an error", address)
	}
}

func TestUDPNetworkAdapter_Dial_SameAddressTwice(t *testing.T) {
	network := adapters.NewUDPNetworkAdapter()

	address := entities.Address{IP: "127.0.0.1", Port: 20020}
	connection1, err := network.Dial(address)
	if err != nil {
		t.Fatalf("Dial(%v) returned unexpected error: %v", address, err)
	}
	defer connection1.Close()

	connection2, err := network.Dial(address)
	if err != nil {
		t.Fatalf("Dial(%v) returned unexpected error: %v", address, err)
	}
	defer connection2.Close()
}

func TestUDPNetworkAdapter_Dial_Send_Receive_Success(t *testing.T) {
	network := adapters.NewUDPNetworkAdapter()

	address := entities.Address{IP: "127.0.0.1", Port: 20021}
	listenConnection, err := network.Listen(address)
	if err != nil {
		t.Fatalf("Listen(%v) returned unexpected error: %v", address, err)
	}
	defer listenConnection.Close()

	dialConnection, err := network.Dial(address)
	if err != nil {
		t.Fatalf("Dial(%v) returned unexpected error: %v", address, err)
	}
	defer dialConnection.Close()

	payload := []byte("hello")
	responsePayload := []byte("response")

	result := make(chan string, 1)
	go func() {
		received, from, err := listenConnection.Receive()
		if err != nil {
			result <- fmt.Sprintf("Receive() returned unexpected error: %v", err)
			return
		}
		if string(received) != string(payload) {
			result <- fmt.Sprintf(
				"Receive() payload = %q; want %q",
				received,
				payload,
			)
			return
		}
		err = listenConnection.SendTo(*from, responsePayload)
		if err != nil {
			result <- fmt.Sprintf(
				"SendTo(%v, %v) returned unexpected error: %v",
				*from,
				responsePayload,
				err,
			)
			return
		}
		result <- ""
	}()

	if err = dialConnection.Send(payload); err != nil {
		t.Fatalf("Send(%q) returned unexpected error: %v", payload, err)
	}

	timeout := uint32(1000)
	received, err := dialConnection.Receive(timeout)
	if err != nil {
		t.Fatalf("Receive(%v) returned unexpected error: %v", timeout, err)
	}
	if string(received) != string(responsePayload) {
		t.Fatalf(
			"Receive(%v) payload = %q; want %q",
			timeout,
			received,
			responsePayload,
		)
	}

	if errorMessage := <-result; errorMessage != "" {
		t.Fatal(errorMessage)
	}
}

func TestUDPNetworkAdapter_Dial_Send_AfterClose(t *testing.T) {
	network := adapters.NewUDPNetworkAdapter()

	address := entities.Address{IP: "127.0.0.1", Port: 20022}
	connection, err := network.Dial(address)
	if err != nil {
		t.Fatalf("Dial(%v) returned unexpected error: %v", address, err)
	}

	if err = connection.Close(); err != nil {
		t.Fatalf("Close() returned unexpected error: %v", err)
	}

	payload := []byte("hello")
	err = connection.Send(payload)
	expectedError := net.ErrClosed
	if !errors.Is(err, expectedError) {
		t.Fatalf("Send(%q) err = %v; want %v", payload, err, expectedError)
	}
}

func TestUDPNetworkAdapter_Dial_Receive_AfterClose(t *testing.T) {
	network := adapters.NewUDPNetworkAdapter()

	address := entities.Address{IP: "127.0.0.1", Port: 20023}
	connection, err := network.Dial(address)
	if err != nil {
		t.Fatalf("Dial(%v) returned unexpected error: %v", address, err)
	}

	if err = connection.Close(); err != nil {
		t.Fatalf("Close() returned unexpected error: %v", err)
	}

	timeout := uint32(0)
	_, err = connection.Receive(timeout)
	expectedError := net.ErrClosed
	if !errors.Is(err, expectedError) {
		t.Fatalf("Receive(%v) err = %v; want %v", timeout, err, expectedError)
	}
}

func TestUDPNetworkAdapter_Dial_Receive_Timeout(t *testing.T) {
	network := adapters.NewUDPNetworkAdapter()

	address := entities.Address{IP: "127.0.0.1", Port: 20024}
	connection, err := network.Dial(address)
	if err != nil {
		t.Fatalf("Dial(%v) returned unexpected error: %v", address, err)
	}
	defer connection.Close()

	timeout := uint32(50)
	result := make(chan error, 1)
	go func() {
		_, err := connection.Receive(timeout)
		result <- err
	}()

	select {
	case err = <-result:
		if err == nil {
			t.Fatalf("Receive(%v) err = nil; want a timeout error", timeout)
		}
		var netErr net.Error
		if !errors.As(err, &netErr) || !netErr.Timeout() {
			t.Fatalf(
				"Receive(%v) err = %v; want a timeout error",
				timeout,
				err,
			)
		}

	case <-time.After(100 * time.Millisecond):
		t.Fatalf("Receive(%v) did not return within 100ms", timeout)
	}
}

func TestUDPNetworkAdapter_Dial_Close_Success(t *testing.T) {
	network := adapters.NewUDPNetworkAdapter()

	address := entities.Address{IP: "127.0.0.1", Port: 20025}
	connection, err := network.Dial(address)
	if err != nil {
		t.Fatalf("Dial(%v) returned unexpected error: %v", address, err)
	}

	if err = connection.Close(); err != nil {
		t.Fatalf("Close() returned unexpected error: %v", err)
	}
}

func TestUDPNetworkAdapter_Dial_Close_AlreadyClosed(t *testing.T) {
	network := adapters.NewUDPNetworkAdapter()

	address := entities.Address{IP: "127.0.0.1", Port: 20026}
	connection, err := network.Dial(address)
	if err != nil {
		t.Fatalf("Dial(%v) returned unexpected error: %v", address, err)
	}

	if err = connection.Close(); err != nil {
		t.Fatalf("Close() returned unexpected error: %v", err)
	}

	err = connection.Close()
	expectedError := net.ErrClosed
	if !errors.Is(err, expectedError) {
		t.Fatalf("Close() err = %v; want %v", err, expectedError)
	}
}
