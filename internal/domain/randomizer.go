package domain

import (
	"math/rand"
)

// Randomizer produces piece types using the official 7-bag algorithm: each
// "bag" contains all seven tetrominoes shuffled, consumed one by one, and the
// bag is refilled (re-shuffled) when empty. This guarantees even piece
// distribution and never emits a piece twice before all seven have appeared.
type Randomizer struct {
	rng  *rand.Rand
	bag  []PieceType
	pos  int
	seed int64
}

// NewRandomizer creates a 7-bag randomizer seeded from the given source.
func NewRandomizer(rng *rand.Rand) *Randomizer {
	r := &Randomizer{rng: rng}
	r.refill()
	return r
}

// NewSeededRandomizer creates a deterministic randomizer for tests/replays.
func NewSeededRandomizer(seed int64) *Randomizer {
	return NewRandomizer(rand.New(rand.NewSource(seed)))
}

func (r *Randomizer) refill() {
	r.bag = AllPieceTypes()
	r.pos = 0
	// Fisher-Yates shuffle.
	for i := len(r.bag) - 1; i > 0; i-- {
		j := r.rng.Intn(i + 1)
		r.bag[i], r.bag[j] = r.bag[j], r.bag[i]
	}
}

// Next returns the next piece type, refilling the bag when exhausted.
func (r *Randomizer) Next() PieceType {
	if r.pos >= len(r.bag) {
		r.refill()
	}
	p := r.bag[r.pos]
	r.pos++
	return p
}

// Peek returns the next n upcoming piece types without consuming them.
func (r *Randomizer) Peek(n int) []PieceType {
	out := make([]PieceType, 0, n)
	// Simulate consumption on a copy of bag state.
	bag := append([]PieceType(nil), r.bag...)
	pos := r.pos
	for len(out) < n {
		if pos >= len(bag) {
			bag = AllPieceTypes()
			for i := len(bag) - 1; i > 0; i-- {
				j := r.rng.Intn(i + 1)
				bag[i], bag[j] = bag[j], bag[i]
			}
			pos = 0
		}
		out = append(out, bag[pos])
		pos++
	}
	return out
}
