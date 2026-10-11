package kademlia_test

import (
	"fmt"
	"sync"
	"testing"

	"kademlia/internal/core/entities"
	"kademlia/internal/core/kademlia"
)

func contactN(n int) kademlia.Contact {
	return kademlia.Contact{ID: entities.NewKademliaIDFromString(fmt.Sprint(n))}
}

func fillBucket(t *testing.T, b *kademlia.Bucket, count int) {
	t.Helper()
	for i := range count {
		if full := b.AddContact(contactN(i)); full != nil {
			t.Fatalf("bucket unexpectedly full at %d", i)
		}
	}
}

func assertOrder(t *testing.T, b *kademlia.Bucket, want ...int) {
	t.Helper()
	got := b.GetContacts()
	if len(got) != len(want) {
		t.Fatalf("got %d contacts, want %d", len(got), len(want))
	}
	for i, n := range want {
		if !got[i].ID.Equals(contactN(n).ID) {
			t.Fatalf("position %d: got %s, want contact %d", i, got[i].ID, n)
		}
	}
}

func TestNewBucket(t *testing.T) {
	b := kademlia.NewBucket()

	if b == nil {
		t.Fatal("NewBucket() returned nil")
	}
	if b.Len() != 0 {
		t.Fatalf("Len() = %d, want 0", b.Len())
	}
	if got := b.GetContacts(); len(got) != 0 {
		t.Fatalf("GetContacts() = %v, want empty", got)
	}
}

func TestBucket_AddContact_AddsToFront(t *testing.T) {
	b := kademlia.NewBucket()

	for i := 1; i <= 3; i++ {
		if got := b.AddContact(contactN(i)); got != nil {
			t.Fatalf("AddContact(%d) = %v, want nil", i, got)
		}
	}

	if b.Len() != 3 {
		t.Fatalf("Len() = %d, want 3", b.Len())
	}
	assertOrder(t, b, 3, 2, 1)
}

func TestBucket_AddContact_ExistingMovesToFront(t *testing.T) {
	b := kademlia.NewBucket()
	for i := 1; i <= 3; i++ {
		b.AddContact(contactN(i))
	}

	if got := b.AddContact(contactN(1)); got != nil {
		t.Fatalf("AddContact(existing) = %v, want nil", got)
	}

	if b.Len() != 3 {
		t.Fatalf("Len() = %d, want 3 (no duplicate)", b.Len())
	}
	assertOrder(t, b, 1, 3, 2)
}

func TestBucket_AddContact_ExistingAlreadyAtFront(t *testing.T) {
	b := kademlia.NewBucket()
	b.AddContact(contactN(1))
	b.AddContact(contactN(2))

	b.AddContact(contactN(2))

	assertOrder(t, b, 2, 1)
}

func TestBucket_AddContact_FullReturnsContact(t *testing.T) {
	b := kademlia.NewBucket()
	fillBucket(t, b, kademlia.BUCKET_SIZE)

	extra := contactN(kademlia.BUCKET_SIZE + 100)
	got := b.AddContact(extra)

	if got == nil {
		t.Fatal("AddContact on full bucket returned nil")
	}
	if !got.ID.Equals(extra.ID) {
		t.Fatalf("returned contact %s, want %s", got.ID, extra.ID)
	}
	if b.Len() != kademlia.BUCKET_SIZE {
		t.Fatalf("Len() = %d, want %d", b.Len(), kademlia.BUCKET_SIZE)
	}
	for _, c := range b.GetContacts() {
		if c.ID.Equals(extra.ID) {
			t.Fatal("rejected contact should not be in the bucket")
		}
	}
}

func TestBucket_AddContact_FullButExistingMovesToFront(t *testing.T) {
	b := kademlia.NewBucket()
	fillBucket(t, b, kademlia.BUCKET_SIZE)

	// Contact 0 was added first, so it is at the back.
	if got := b.AddContact(contactN(0)); got != nil {
		t.Fatalf("AddContact(existing) on full bucket = %v, want nil", got)
	}

	if b.Len() != kademlia.BUCKET_SIZE {
		t.Fatalf("Len() = %d, want %d", b.Len(), kademlia.BUCKET_SIZE)
	}
	if first := b.GetContacts()[0]; !first.ID.Equals(contactN(0).ID) {
		t.Fatalf("front = %s, want contact 0", first.ID)
	}
}

func TestBucket_RemoveContact(t *testing.T) {
	tests := []struct {
		name   string
		remove int
		want   []int
	}{
		{"front", 3, []int{2, 1}},
		{"middle", 2, []int{3, 1}},
		{"back", 1, []int{3, 2}},
		{"missing", 99, []int{3, 2, 1}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := kademlia.NewBucket()
			for i := 1; i <= 3; i++ {
				b.AddContact(contactN(i))
			}

			b.RemoveContact(contactN(tt.remove))

			if b.Len() != len(tt.want) {
				t.Fatalf("Len() = %d, want %d", b.Len(), len(tt.want))
			}
			assertOrder(t, b, tt.want...)
		})
	}
}

func TestBucket_RemoveContact_EmptyBucket(t *testing.T) {
	b := kademlia.NewBucket()

	b.RemoveContact(contactN(1))

	if b.Len() != 0 {
		t.Fatalf("Len() = %d, want 0", b.Len())
	}
}

func TestBucket_RemoveContact_FreesSlotWhenFull(t *testing.T) {
	b := kademlia.NewBucket()
	fillBucket(t, b, kademlia.BUCKET_SIZE)

	b.RemoveContact(contactN(0))

	if got := b.AddContact(contactN(1000)); got != nil {
		t.Fatalf("AddContact after removal = %v, want nil", got)
	}
	if b.Len() != kademlia.BUCKET_SIZE {
		t.Fatalf("Len() = %d, want %d", b.Len(), kademlia.BUCKET_SIZE)
	}
}

func TestBucket_GetContactAndCalcDistance(t *testing.T) {
	b := kademlia.NewBucket()
	for i := 1; i <= 3; i++ {
		b.AddContact(contactN(i))
	}
	target := entities.NewKademliaIDFromString("target")

	got := b.GetContactAndCalcDistance(target)

	if len(got) != 3 {
		t.Fatalf("got %d contacts, want 3", len(got))
	}
	for i, n := range []int{3, 2, 1} {
		if !got[i].ID.Equals(contactN(n).ID) {
			t.Fatalf("position %d: got %s, want contact %d", i, got[i].ID, n)
		}
	}
	if b.Len() != 3 {
		t.Fatalf("bucket should be unchanged, Len() = %d", b.Len())
	}
}

func TestBucket_GetContactAndCalcDistance_Empty(t *testing.T) {
	b := kademlia.NewBucket()

	got := b.GetContactAndCalcDistance(entities.NewKademliaIDFromString("target"))

	if len(got) != 0 {
		t.Fatalf("got %v, want empty", got)
	}
}

func TestBucket_GetContacts_ReturnsCopy(t *testing.T) {
	b := kademlia.NewBucket()
	b.AddContact(contactN(1))
	b.AddContact(contactN(2))

	got := b.GetContacts()
	got[0] = contactN(99)

	assertOrder(t, b, 2, 1)
}

func TestBucket_Len(t *testing.T) {
	b := kademlia.NewBucket()

	for i := 1; i <= 3; i++ {
		b.AddContact(contactN(i))
		if b.Len() != i {
			t.Fatalf("Len() = %d, want %d", b.Len(), i)
		}
	}

	b.AddContact(contactN(1)) // duplicate
	if b.Len() != 3 {
		t.Fatalf("Len() after duplicate = %d, want 3", b.Len())
	}

	b.RemoveContact(contactN(1))
	if b.Len() != 2 {
		t.Fatalf("Len() after remove = %d, want 2", b.Len())
	}
}

func TestBucket_ConcurrentAccess(t *testing.T) {
	b := kademlia.NewBucket()
	target := entities.NewKademliaIDFromString("target")
	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(4)
		go func(i int) { defer wg.Done(); b.AddContact(contactN(i)) }(i)
		go func(i int) { defer wg.Done(); b.RemoveContact(contactN(i)) }(i)
		go func() { defer wg.Done(); b.GetContacts() }()
		go func() { defer wg.Done(); b.GetContactAndCalcDistance(target) }()
	}
	wg.Wait()

	if b.Len() > kademlia.BUCKET_SIZE {
		t.Fatalf("Len() = %d exceeds BUCKET_SIZE", b.Len())
	}
}
