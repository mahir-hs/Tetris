package domain

import (
	"testing"
)

func TestRandomizerBagContainsAllSeven(t *testing.T) {
	r := NewSeededRandomizer(42)
	seen := make(map[PieceType]int)
	for i := 0; i < 7; i++ {
		seen[r.Next()]++
	}
	if len(seen) != 7 {
		t.Fatalf("expected all 7 piece types in one bag, got %d", len(seen))
	}
	for _, p := range AllPieceTypes() {
		if seen[p] != 1 {
			t.Errorf("piece %d appeared %d times in a single bag, want 1", p, seen[p])
		}
	}
}

func TestRandomizerNoRepeatsWithinBag(t *testing.T) {
	r := NewSeededRandomizer(7)
	counts := make(map[PieceType]int)
	for i := 0; i < 14; i++ { // two full bags
		counts[r.Next()]++
	}
	for _, p := range AllPieceTypes() {
		if counts[p] != 2 {
			t.Errorf("piece %d appeared %d times over two bags, want 2", p, counts[p])
		}
	}
}

func TestRandomizerPeekDoesNotConsume(t *testing.T) {
	r := NewSeededRandomizer(123)
	peeked := r.Peek(5)
	if len(peeked) != 5 {
		t.Fatalf("peek returned %d pieces, want 5", len(peeked))
	}
	// Consuming 5 pieces should match the peek exactly.
	for i, p := range peeked {
		if got := r.Next(); got != p {
			t.Errorf("after peek, piece %d = %v, want %v", i, got, p)
		}
	}
}

func TestRandomizerDeterministicWithSeed(t *testing.T) {
	a := NewSeededRandomizer(99)
	b := NewSeededRandomizer(99)
	for i := 0; i < 21; i++ {
		if a.Next() != b.Next() {
			t.Fatalf("seeded randomizers diverged at draw %d", i)
		}
	}
}
