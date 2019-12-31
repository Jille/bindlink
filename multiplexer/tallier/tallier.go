package tallier

import (
	"time"
)

type tallierBucket struct {
	start int64
	count uint64
}

type Tallier struct {
	bucketSize int64
	window     int64
	nBuckets   int64
	buckets    []tallierBucket
}

func nowMillis() int64 {
	return time.Now().UnixNano() / 1000000
}

func New(bucketSize, window int64) *Tallier {
	nBuckets := window / bucketSize
	return &Tallier{
		bucketSize: bucketSize,
		window:     window,
		nBuckets:   nBuckets,
		buckets:    make([]tallierBucket, nBuckets),
	}
}

func (t *Tallier) Tally() {
	t.TallyN(1)
}

func (t *Tallier) TallyN(n uint64) {
	start := nowMillis() / t.bucketSize
	bucket := start % t.nBuckets
	if t.buckets[bucket].start != start {
		t.buckets[bucket].start = start
		t.buckets[bucket].count = 0
	}
	t.buckets[bucket].count += n
}

// Returns the number of tallies in the last window milliseconds.
func (t *Tallier) Count() uint64 {
	ret := uint64(0)
	start := (nowMillis() - t.window) / t.bucketSize
	for i := 0; i < int(t.nBuckets); i++ {
		if t.buckets[i].start < start {
			continue
		}
		ret += t.buckets[i].count
	}
	return ret
}
