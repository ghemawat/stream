package pipe

import (
	"math/rand"
	"time"
)

// Sample picks n pseudo-randomly chosen input items.
func Sample(n int) Filter {
	return func(arg Arg) {
		// Could speed this up by using Algorithm Z from Vitter.
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		reservoir := make([]string, 0, n)
		i := 0
		for s := range arg.In {
			switch {
			case i < n:
				reservoir = append(reservoir, s)
			case r.Float32() < float32(n)/float32(i+1):
				reservoir[r.Intn(n)] = s
			}
			i++
		}
		for _, s := range reservoir {
			arg.Out <- s
		}
	}
}
