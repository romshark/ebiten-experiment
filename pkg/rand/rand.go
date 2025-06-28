package rand

import (
	"fmt"
	"time"

	"math/rand/v2"
)

type RNG struct {
	src rand.Rand
}

func New(seed1, seed2 uint64) *RNG {
	return &RNG{
		src: *rand.New(rand.NewPCG(seed1, seed2)),
	}
}

// NewAuto creates an RNG with a time-based seed.
func NewAuto() *RNG {
	u := uint64(time.Now().UnixNano())
	return New(u, u+1)
}

// Int returns a random int in [0, max).
func (r *RNG) Int(max int) int {
	return r.src.IntN(max)
}

// IntRange returns a random int in [min, max).
func (r *RNG) IntRange(min, max int) int {
	if max <= min {
		panic(fmt.Errorf("invalid range: %d-%d", min, max))
	}
	return min + r.src.IntN(max-min)
}
