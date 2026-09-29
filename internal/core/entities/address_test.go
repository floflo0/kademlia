package entities_test

import (
	"kademlia/internal/core/entities"
	"reflect"
	"testing"
)

func TestNewAddressFromString_Success(t *testing.T) {
	tests := []struct {
		addressString   string
		expectedAddress *entities.Address
	}{
		{
			"127.0.0.1:8080",
			&entities.Address{IP: "127.0.0.1", Port: 8080},
		},
		{
			"172.30.1.2:12",
			&entities.Address{IP: "172.30.1.2", Port: 12},
		},
		{
			"0.0.0.0:3000",
			&entities.Address{IP: "0.0.0.0", Port: 3000},
		},
	}

	for _, test := range tests {
		address, err := entities.NewAddressFromString(test.addressString)
		if err != nil {
			t.Fatalf(
				"NewAddressFromString(%q) returned unexpected error: %v",
				test.addressString,
				err,
			)
		}
		if !reflect.DeepEqual(address, test.expectedAddress) {
			t.Fatalf(
				"NewAddressFromString(%q) address = %v; want %v",
				test.addressString,
				address,
				test.expectedAddress,
			)
		}
	}
}

func TestNewAddressFromString_NoPort(t *testing.T) {
	addressString := "127.0.0.1"
	_, err := entities.NewAddressFromString(addressString)
	if err == nil {
		t.Fatalf(
			"NewAddressFromString(%q)  expected error, got nil",
			addressString,
		)
	}
}

func TestNewAddressFromString_InvalidPort(t *testing.T) {
	addressString := "127.0.0.1:invalid"
	_, err := entities.NewAddressFromString(addressString)
	if err == nil {
		t.Fatalf(
			"NewAddressFromString(%q)  expected error, got nil",
			addressString,
		)
	}
}

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
