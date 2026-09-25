package kademlia

import (
	"kademlia/internal/core/entities"
	"time"
)

type mockKademlia struct {
	runFunction           func(firstContact *entities.Address) error
	pingFunction          func(address entities.Address) (time.Duration, error)
	getBucketsFunction    func() []*Bucket
	getStoredKeysFunction func() []string
	quitFunction          func() error
}

func NewMockKademlia(
	runFunction func(firstContact *entities.Address) error,
	pingFunction func(address entities.Address) (time.Duration, error),
	getBucketsFunction func() []*Bucket,
	getStoredKeysFunction func() []string,
	quitFunction func() error,
) Kademlia {
	return &mockKademlia{
		runFunction:           runFunction,
		pingFunction:          pingFunction,
		getBucketsFunction:    getBucketsFunction,
		getStoredKeysFunction: getStoredKeysFunction,
		quitFunction:          quitFunction,
	}
}

func (k *mockKademlia) Run(firstContact *entities.Address) error {
	return k.runFunction(firstContact)
}

func (k *mockKademlia) Ping(address entities.Address) (time.Duration, error) {
	return k.pingFunction(address)
}

func (k *mockKademlia) GetBuckets() []*Bucket {
	return k.getBucketsFunction()
}

func (k *mockKademlia) GetStoredKeys() []string {
	return k.getStoredKeysFunction()
}

func (k *mockKademlia) Quit() error {
	return k.quitFunction()
}
