package entities_test

import (
	"strings"
	"testing"

	"kademlia/internal/core/entities"
)

func TestNewKademliaID(t *testing.T) {
	valid := strings.Repeat("ab", entities.KademliaIDLength)

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid", valid, false},
		{"longer than needed is truncated", valid + "ff", false},
		{"invalid hex", strings.Repeat("zz", entities.KademliaIDLength), true},
		{"odd length", "abc", true},
		{"too short", "abcd", true},
		{"empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := entities.NewKademliaID(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error for %q", tt.input)
				}
				if id != nil {
					t.Fatalf("expected nil ID on error, got %v", id)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if id.String() != valid {
				t.Fatalf("got %s, want %s", id.String(), valid)
			}
		})
	}
}

func TestNewRandomKademliaID(t *testing.T) {
	a := entities.NewRandomKademliaID()
	b := entities.NewRandomKademliaID()

	if a == nil || b == nil {
		t.Fatal("expected non-nil IDs")
	}
	if a.Equals(b) {
		t.Fatal("two random IDs should not be equal")
	}
}

func TestNewKademliaIDFromAddress(t *testing.T) {
	addr := entities.Address{IP: "127.0.0.1", Port: 8000}

	got := entities.NewKademliaIDFromAddress(addr)
	want := entities.NewKademliaIDFromString("127.0.0.18000")

	if !got.Equals(want) {
		t.Fatalf("got %s, want %s", got, want)
	}

	other := entities.NewKademliaIDFromAddress(entities.Address{IP: "127.0.0.1", Port: 8001})
	if got.Equals(other) {
		t.Fatal("different ports should give different IDs")
	}
}

func mustID(t *testing.T, hexStr string) *entities.KademliaID {
	t.Helper()
	id, err := entities.NewKademliaID(hexStr)
	if err != nil {
		t.Fatalf("NewKademliaID(%q): %v", hexStr, err)
	}
	return id
}

func TestKademliaID_Less(t *testing.T) {
	zero := strings.Repeat("00", entities.KademliaIDLength)
	lowLast := strings.Repeat("00", entities.KademliaIDLength-1) + "01"
	highFirst := "01" + strings.Repeat("00", entities.KademliaIDLength-1)

	tests := []struct {
		name string
		a, b string
		want bool
	}{
		{"equal", zero, zero, false},
		{"less on first byte", zero, highFirst, true},
		{"greater on first byte", highFirst, zero, false},
		{"less on last byte", zero, lowLast, true},
		{"greater on last byte", lowLast, zero, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, b := mustID(t, tt.a), mustID(t, tt.b)
			if got := a.Less(b); got != tt.want {
				t.Fatalf("Less() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestKademliaID_Equals(t *testing.T) {
	a := mustID(t, strings.Repeat("ab", entities.KademliaIDLength))
	same := mustID(t, strings.Repeat("ab", entities.KademliaIDLength))
	diffLast := mustID(t, strings.Repeat("ab", entities.KademliaIDLength-1)+"ac")
	diffFirst := mustID(t, "ac"+strings.Repeat("ab", entities.KademliaIDLength-1))

	if !a.Equals(same) {
		t.Fatal("identical IDs should be equal")
	}
	if a.Equals(diffLast) {
		t.Fatal("IDs differing on last byte should not be equal")
	}
	if a.Equals(diffFirst) {
		t.Fatal("IDs differing on first byte should not be equal")
	}
}

func TestKademliaID_CalcDistance(t *testing.T) {
	a := mustID(t, strings.Repeat("ff", entities.KademliaIDLength))
	b := mustID(t, strings.Repeat("0f", entities.KademliaIDLength))
	zero := mustID(t, strings.Repeat("00", entities.KademliaIDLength))

	if got := a.CalcDistance(a); !got.Equals(zero) {
		t.Fatalf("distance to self = %s, want zero", got)
	}

	want := mustID(t, strings.Repeat("f0", entities.KademliaIDLength))
	if got := a.CalcDistance(b); !got.Equals(want) {
		t.Fatalf("got %s, want %s", got, want)
	}

	if !a.CalcDistance(b).Equals(b.CalcDistance(a)) {
		t.Fatal("distance should be symmetric")
	}
}

func TestKademliaID_String(t *testing.T) {
	hexStr := "2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae"
	id := mustID(t, hexStr)

	if got := id.String(); got != hexStr {
		t.Fatalf("String() = %q, want %q", got, hexStr)
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
		id, _ := entities.NewKademliaID(test.id)
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
