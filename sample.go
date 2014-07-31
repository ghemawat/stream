package stream

import (
	"math/rand"
	"time"
)

// Sample picks n pseudo-randomly chosen input items.
// TODO: Maybe add SampleWithSeed.
func Sample(n int) Filter {
	return FilterFunc(func(arg Arg) error {
		// Could speed this up by using Algorithm Z from Vitter.
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		reservoir := make([]string, 0, n)
		i := 0
		for s := range arg.In {
			if i < n {
				reservoir = append(reservoir, s)
			} else {
				j := r.Intn(i + 1)
				if j < n {
					reservoir[j] = s
				}
			}
			i++
		}
		for _, s := range reservoir {
			arg.Out <- s
		}
		return nil
	})
}
