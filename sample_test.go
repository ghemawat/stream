package stream_test

import (
	"github.com/ghemawat/stream"

	"fmt"
	"testing"
)

// doTest checks that Sample picks items evenly.  "n" samples are
// drawn from a list of numbers of length "space".  This is repeated
// "iters" times.  The number of times a particular number is drawn
// should be within "tolerance" of the expected number.
func doTest(t *testing.T, n, space, iters int, tolerance float64) {
	count := make([]int, space)
	for i := 0; i < iters; i++ {
		s := stream.Sequence(
			stream.Numbers(0, space-1),
			stream.SampleWithSeed(n, int64(i)),
		)
		stream.ForEach(s, func(s string) {
			num := -1 // Will cause panic below if Scan fails
			fmt.Sscan(s, &num)
			count[num]++
		})
	}

	// Check that all counts are approximately equal.
	expected := (float64(iters) * float64(n)) / float64(space)
	minExpected := expected * (1.0 - tolerance)
	maxExpected := expected * (1.0 + tolerance)
	for i, n := range count {
		if float64(n) < minExpected || float64(n) > maxExpected {
			t.Errorf("%d has %d samples; expected range [%f,%f]\n",
				i, n, minExpected, maxExpected)
		}
	}
}

func TestSample_1of2(t *testing.T)    { doTest(t, 1, 2, 10000, 0.01) }
func TestSample_9of10(t *testing.T)   { doTest(t, 9, 10, 1000, 0.05) }
func TestSample_50of100(t *testing.T) { doTest(t, 50, 100, 1000, 0.15) }
func TestSample_99of100(t *testing.T) { doTest(t, 99, 100, 100, 0.05) }
