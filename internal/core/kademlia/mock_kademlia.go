package kademlia

import (
	"kademlia/internal/core/entities"
	"time"
)

type MockKademlia struct {
	RunFunction           func(firstContact *entities.Address) error
	PingFunction          func(address entities.Address) (time.Duration, error)
	GetBucketsFunction    func() []*bucket
	GetStoredKeysFunction func() []string
}

func (k *MockKademlia) Run(firstContact *entities.Address) error {
	if k.RunFunction != nil {
		return k.RunFunction(firstContact)
	}
	return nil
}

func (k *MockKademlia) Ping(address entities.Address) (time.Duration, error) {
	if k.PingFunction != nil {
		return k.PingFunction(address)
	}
	return 0, nil
}

func (k *MockKademlia) GetBuckets() []*bucket {
	if k.GetBucketsFunction != nil {
		return k.GetBucketsFunction()
	}
	return []*bucket{}
}

func (k *MockKademlia) GetStoredKeys() []string {
	if k.GetStoredKeysFunction != nil {
		return k.GetStoredKeysFunction()
	}
	return []string{}
}
