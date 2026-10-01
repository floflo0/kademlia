package entities

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math/rand"
)

// IDLength is the static number of bytes in a KademliaID.
const KademliaIDLength = 32 // 256 bit / 8 bits/byte = 32 bytes

type KademliaID [KademliaIDLength]byte

// NewKademliaID returns a new instance of a KademliaID based on the string input
func NewKademliaID(data string) (*KademliaID, error) {
	decoded, err := hex.DecodeString(data)
	if err != nil {
		return nil, err
	}
	if len(decoded) < KademliaIDLength {
		return nil, errors.New("ID too short")
	}

	newKademliaID := KademliaID{}
	for i := range KademliaIDLength {
		newKademliaID[i] = decoded[i]
	}

	return &newKademliaID, nil
}

// NewRandomKademliaID returns a new instance of a random KademliaID,
// change this to a better version if you like
func NewRandomKademliaID() *KademliaID {
	newKademliaID := KademliaID{}
	for i := range KademliaIDLength {
		newKademliaID[i] = uint8(rand.Intn(256))
	}
	return &newKademliaID
}

func NewKademliaIDFromString(value string) *KademliaID {
	hash := sha256.Sum256([]byte(value))
	return (*KademliaID)(&hash)
}

func NewKademliaIDFromAddress(address Address) *KademliaID {
	return NewKademliaIDFromString(address.IP + fmt.Sprint(address.Port))
}

// Less returns true if id < otherKademliaID (bitwise)
func (id KademliaID) Less(otherKademliaID *KademliaID) bool {
	for i := range KademliaIDLength {
		if id[i] != otherKademliaID[i] {
			return id[i] < otherKademliaID[i]
		}
	}
	return false
}

// Equals returns true if id == otherKademliaID (bitwise)
func (id KademliaID) Equals(otherKademliaID *KademliaID) bool {
	for i := range KademliaIDLength {
		if id[i] != otherKademliaID[i] {
			return false
		}
	}
	return true
}

// CalcDistance returns a new instance of a KademliaID that is built
// through a bitwise XOR operation betweeen kademliaID and target
func (id KademliaID) CalcDistance(target *KademliaID) *KademliaID {
	result := KademliaID{}
	for i := range KademliaIDLength {
		result[i] = id[i] ^ target[i]
	}
	return &result
}

// String returns a simple string representation of a KademliaID
func (id *KademliaID) String() string {
	return hex.EncodeToString(id[0:KademliaIDLength])
}

func (id *KademliaID) ShortString() string {
	return fmt.Sprintf("%04x...%04x", id[:4], id[KademliaIDLength-4:])
}
