package kademlia

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"sync"
	"testing"
)

func TestDataStore_PutGet_Success(t *testing.T) {
	ds := NewDataStore()
	value := "Hola Kademlia"

	hash := sha256.Sum256([]byte(value))
	key := hex.EncodeToString(hash[:])

	// Put test
	err := ds.Put(key, value)
	if err != nil {
		t.Fatalf("We expected success, but it failed with: %v", err)
	}

	// Get test
	got, err := ds.Get(key)
	if err != nil {
		t.Fatalf("We expected success, but it failed with: %v", err)
	}

	if got != value {
		t.Errorf("We obtained %s but we were expecting %s", got, value)
	}
}

func TestDataStore_Put_InvalidHash(t *testing.T) {
	ds := NewDataStore()
	value := "Hola Kademlia"
	fakeKey := "0123456789abcdef0123456789abcdef0123456789abcdef"

	err := ds.Put(fakeKey, value)
	if !errors.Is(err, ErrHashMismatch) {
		t.Fatalf("We were expecting the error errHashMismatch, but we obtained %v", err)
	}
}

func TestDataStore_Get_NotFound(t *testing.T) {
	ds := NewDataStore()
	fakeKey := "0123456789abcdef0123456789abcdef0123456789abcdef"

	_, err := ds.Get(fakeKey)
	if !errors.Is(err, ErrKeyNotFound) {
		t.Fatalf("We were expecting the error errNotFound, but we obtained %v", err)
	}
}

func TestDataStore_Concurrency(t *testing.T) {
	ds := NewDataStore()
	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(val string) {
			defer wg.Done()
			hash := sha256.Sum256([]byte(val))
			key := hex.EncodeToString(hash[:])
			_ = ds.Put(key, val)
		}(string(rune(i)))
	}

	wg.Wait()

	if len(ds.Keys()) == 0 {
		t.Error("We were expecting saved keys after the concurrent execution.")
	}
}
