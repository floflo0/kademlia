package kademlia_test

import (
	"kademlia/internal/core/kademlia"
	"testing"
)

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
		id := kademlia.NewKademliaID(test.id)
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
