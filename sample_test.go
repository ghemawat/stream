package stream_test

import (
	"github.com/ghemawat/stream"

	"fmt"
	"testing"
)

func TestSample(t *testing.T) {
	// A weak test that Sample picks items evenly.
	const space = 100  // Space of numbers to sample from
	const samples = 50 // Number of samples drawn per run
	const iters = 1000 // Number of runs

	var count [space]int
	for i := 0; i < iters; i++ {
		s := stream.Sequence(
			stream.Numbers(0, space-1),
			stream.SampleWithSeed(samples, int64(i)),
		)
		stream.ForEach(s, func(s string) {
			num := -1 // Will cause panic below if Scan fails
			fmt.Sscan(s, &num)
			count[num]++
		})
	}

	// Check that all counts are approximately equal.
	const expected = (iters * samples) / space
	const minExpected = expected * 0.85
	const maxExpected = expected * 1.15
	for i, n := range count {
		if n < minExpected || n > maxExpected {
			t.Errorf("%d has %d samples; expected range [%f,%f]\n",
				i, n, minExpected, maxExpected)
		}
	}
}