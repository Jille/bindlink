package sampler

import (
	"math/rand"
)

type Sampler struct {
	sum     float64
	offsets []float64
	keyLut  []int
}

func New(weights map[int]float64) *Sampler {
	sum := float64(0)
	keyLut := []int{}
	for key, _ := range weights {
		keyLut = append(keyLut, key)
	}
	for _, w := range weights {
		sum += w
	}
	prev := float64(0)
	offsets := []float64{}
	for _, key := range keyLut {
		offset := prev + weights[key]
		offsets = append(offsets, offset)
		prev = offset
	}
	return &Sampler{
		sum:     sum,
		offsets: offsets,
		keyLut:  keyLut,
	}
}

func (s *Sampler) Sample() int {
	v := rand.Float64() * s.sum
	// TODO bisect.
	for i, offset := range s.offsets {
		if offset > v {
			return s.keyLut[i]
		}
	}
	return s.keyLut[0]
}
