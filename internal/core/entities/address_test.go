package entities_test

import (
	"kademlia/internal/core/entities"
	"testing"
)

func TestAddress_String(t *testing.T) {
	tests := []struct {
		address        entities.Address
		expectedString string
	}{
		{
			entities.Address{IP: "127.0.0.1", Port: 8080},
			"127.0.0.1:8080",
		},
		{
			entities.Address{IP: "172.30.1.2", Port: 12},
			"172.30.1.2:12",
		},
		{
			entities.Address{IP: "0.0.0.0", Port: 3000},
			"0.0.0.0:3000",
		},
	}

	for _, test := range tests {
		string := test.address.String()
		if string != test.expectedString {
			t.Fatalf(
				"String() string = %q; want %q",
				string,
				test.expectedString,
			)
		}
	}
}
