package kademlia_test

import (
	"kademlia/internal/core/kademlia"
	"testing"
)

func TestNewKademliaIDFromString(t *testing.T) {
	hash1 := "2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae"
	hash2 := "fcde2b2edba56bf408601fb721fe9b5c338d10ee429ea04fae5511b68fbf8fb9"
	hash3 := "baa5a0964d3320fbc0c6a922140453c8513ea24ab8fd0577034804a967248096"
	expectedID1, err := kademlia.NewKademliaID(hash1)
	if err != nil {
		t.Fatalf("NewKademliaID(%q) returned unexpected error: %v", hash1, err)
	}
	expectedID2, err := kademlia.NewKademliaID(hash2)
	if err != nil {
		t.Fatalf("NewKademliaID(%q) returned unexpected error: %v", hash2, err)
	}
	expectedID3, err := kademlia.NewKademliaID(hash3)
	if err != nil {
		t.Fatalf("NewKademliaID(%q) returned unexpected error: %v", hash3, err)
	}

	tests := []struct {
		value      string
		expectedID *kademlia.KademliaID
	}{
		{"foo", expectedID1},
		{"bar", expectedID2},
		{"baz", expectedID3},
	}

	for _, test := range tests {
		id := kademlia.NewKademliaIDFomString(test.value)
		if !id.Equals(test.expectedID) {
			t.Fatalf(
				"NewKademliaIDFomString() id = %q; want %q",
				id.String(),
				test.expectedID,
			)
		}
	}
}

func TestKademliaID_ShortString(t *testing.T) {
	tests := []struct {
		id             string
		expectedString string
	}{
		{
			"0000000000000000000000000000000000000000000000000000000000000000",
			"00000000...00000000",
		},
		{
			"ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
			"ffffffff...ffffffff",
		},
		{
			"0123456700000000000000000000000000000000000000000000000089abcdef",
			"01234567...89abcdef",
		},
	}

	for _, test := range tests {
		id, _ := kademlia.NewKademliaID(test.id)
		string := id.ShortString()
		if string != test.expectedString {
			t.Fatalf(
				"ShortString() string = %q; want %q",
				string,
				test.expectedString,
			)
		}
	}
}
