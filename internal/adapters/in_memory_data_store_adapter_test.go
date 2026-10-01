package adapters_test

import (
	"errors"
	"fmt"
	"kademlia/internal/adapters"
	"kademlia/internal/core/entities"
	"kademlia/internal/core/ports"
	"sync"
	"testing"
)

func TestNewInMemoryDataStore(t *testing.T) {
	dataStore := adapters.NewInMemoryDataStore()
	if dataStore == nil {
		t.Fatal("NewInMemoryDataStore() returns nil")
	}
}

func TestInMemoryDataStore_Put_Success(t *testing.T) {
	dataStore := adapters.NewInMemoryDataStore()

	value := "Hello, World!"
	key := entities.NewKademliaIDFromString(value)

	err := dataStore.Put(*key, value)
	if err != nil {
		t.Errorf(
			"Put(%q, %q) returns unexpected error: %v",
			key.String(),
			value,
			err,
		)
	}
}

func TestInMemoryDataStore_Put_InvalidHash(t *testing.T) {
	dataStore := adapters.NewInMemoryDataStore()
	value := "Hello, World!"
	invalidKey, err := entities.NewKademliaID(
		"0000000000000000000000000000000000000000000000000000000000000000",
	)
	if err != nil {
		t.Fatalf("NewKademliaID() returns unexpected error: %v", err)
	}

	err = dataStore.Put(*invalidKey, value)
	expectedError := ports.ErrHashMismatch
	if !errors.Is(err, expectedError) {
		t.Fatalf(
			"Put(%q, %q) err = %v; want %v",
			invalidKey.String(),
			value,
			err,
			expectedError,
		)
	}
}

func TestInMemoryDataStore_Get_NotFound(t *testing.T) {
	dataStore := adapters.NewInMemoryDataStore()
	key, err := entities.NewKademliaID(
		"0000000000000000000000000000000000000000000000000000000000000000",
	)
	if err != nil {
		t.Fatalf("NewKademliaID() returns unexpected error: %v", err)
	}

	_, err = dataStore.Get(*key)
	expectedError := ports.ErrKeyNotFound
	if !errors.Is(err, expectedError) {
		t.Fatalf(
			"Get(%q) err = %v; want %v",
			key.String(),
			err,
			expectedError,
		)
	}
}

func TestInMemoryDataStore_Gut_Success(t *testing.T) {
	dataStore := adapters.NewInMemoryDataStore()

	expectedValue := "Hello, World!"
	key, err := entities.NewKademliaID(
		"dffd6021bb2bd5b0af676290809ec3a53191dd81c7f70a4b28688a362182986f",
	)
	if err != nil {
		t.Fatalf("NewKademliaID() returns unexpected error: %v", err)
	}

	err = dataStore.Put(*key, expectedValue)
	if err != nil {
		t.Errorf(
			"Put(%q, %q) returns unexpected error: %v",
			key.String(),
			expectedValue,
			err,
		)
	}

	value, err := dataStore.Get(*key)
	if err != nil {
		t.Errorf("Get(%q) returns unexpected error: %v", key.String(), err)
	}
	if value != expectedValue {
		t.Errorf(
			"Get(%q) value = %q; want %q",
			key.String(),
			value,
			expectedValue,
		)
	}
}

func TestInMemoryDataStore_Concurrency(t *testing.T) {
	dataStore := adapters.NewInMemoryDataStore()

	count := 100
	var waitGroup sync.WaitGroup
	for i := range count {
		waitGroup.Go(func() {
			value := fmt.Sprintf("data %d", i)
			key := entities.NewKademliaIDFromString(value)
			err := dataStore.Put(*key, value)
			if err != nil {
				t.Errorf(
					"Put(%q, %q) returns unexpected error: %v",
					key.String(),
					value,
					err,
				)
			}
		})
	}
	waitGroup.Wait()

	keys := dataStore.Keys()
	keysCount := len(keys)
	if keysCount != count {
		t.Fatalf(
			"Keys() returned %d keys, want %d",
			keysCount,
			count,
		)
	}

	for _, key := range keys {
		waitGroup.Go(func() {
			_, err := dataStore.Get(key)
			if err != nil {
				t.Errorf(
					"Get(%q) returns unexpected error: %v",
					key.String(),
					err,
				)
			}
		})
	}
	waitGroup.Wait()
}
