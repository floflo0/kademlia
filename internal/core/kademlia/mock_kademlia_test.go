package kademlia_test

import (
	"errors"
	"kademlia/internal/core/entities"
	"kademlia/internal/core/kademlia"
	"testing"
	"time"
)

func TestMockKademlia_Run_NilFunction(t *testing.T) {
	mockKademlia := &kademlia.MockKademlia{}
	err := mockKademlia.Run()
	if err != nil {
		t.Fatalf("Run() returned unexpected error: %v", err)
	}
}

func TestMockKademlia_Run_CallsFunction(t *testing.T) {
	called := false
	mockKademlia := &kademlia.MockKademlia{
		RunFunction: func() error {
			called = true
			return nil
		},
	}
	err := mockKademlia.Run()
	if err != nil {
		t.Fatalf("Run() returned unexpected error: %v", err)
	}
	if !called {
		t.Fatal("Run() didn't call RunFunction")
	}
}

func TestMockKademlia_Run_ReturnsError(t *testing.T) {
	called := false
	expectedError := errors.New("test error")
	mockKademlia := &kademlia.MockKademlia{
		RunFunction: func() error {
			called = true
			return expectedError
		},
	}
	err := mockKademlia.Run()
	if !errors.Is(err, expectedError) {
		t.Fatalf("Run() err = %v; want %v", err, expectedError)
	}
	if !called {
		t.Fatal("Run() didn't call RunFunction")
	}
}

func TestMockKademlia_Ping_NilFunction(t *testing.T) {
	mockKademlia := &kademlia.MockKademlia{}
	address := entities.Address{IP: "127.0.0.1", Port: 8080}

	elapsedTime, err := mockKademlia.Ping(address)
	if err != nil {
		t.Fatalf(
			"Ping(%v) with nil PingFunction returned unexpected error: %v",
			address,
			err,
		)
	}
	expectedElapsedTime := time.Duration(0)
	if elapsedTime != expectedElapsedTime {
		t.Fatalf(
			"Ping(%v) with nil PingFunction returned elapsedTime = %v; want %v",
			address,
			elapsedTime,
			expectedElapsedTime,
		)
	}
}

func TestMockKademlia_Ping_CallsFunction(t *testing.T) {
	called := false
	var gotAddress entities.Address
	expectedElapsedTime := 5 * time.Millisecond
	mockKademlia := &kademlia.MockKademlia{
		PingFunction: func(address entities.Address) (time.Duration, error) {
			called = true
			gotAddress = address
			return expectedElapsedTime, nil
		},
	}
	address := entities.Address{IP: "127.0.0.1", Port: 8080}

	elapsedTime, err := mockKademlia.Ping(address)
	if err != nil {
		t.Fatalf("Ping(%v) returned unexpected error: %v", address, err)
	}
	if !called {
		t.Fatalf("Ping(%v) didn't call PingFunction", address)
	}
	if gotAddress != address {
		t.Fatalf(
			"PingFunction() called with address = %v; want %v",
			gotAddress,
			address,
		)
	}
	if elapsedTime != expectedElapsedTime {
		t.Fatalf(
			"Ping(%v) elapsedTime = %v; want %v",
			address,
			elapsedTime,
			expectedElapsedTime,
		)
	}
}

func TestMockKademlia_Ping_ReturnsError(t *testing.T) {
	expectedError := errors.New("test error")
	mock := &kademlia.MockKademlia{
		PingFunction: func(address entities.Address) (time.Duration, error) {
			return 0, expectedError
		},
	}
	address := entities.Address{IP: "127.0.0.1", Port: 8080}

	_, err := mock.Ping(address)
	if !errors.Is(err, expectedError) {
		t.Fatalf("Ping(%v) err = %v; want %v", address, err, expectedError)
	}
}
