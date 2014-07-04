package pipe

import (
	"container/heap"
	"crypto/sha1"
	"fmt"
)

type sample struct {
	hash string
	item string
}

type samples []sample

func (h samples) Len() int            { return len(h) }
func (h samples) Less(i, j int) bool  { return h[i].hash < h[j].hash }
func (h samples) Swap(i, j int)       { h[i], h[j] = h[j], h[i] }
func (h *samples) Push(x interface{}) { *h = append(*h, x.(sample)) }
func (h *samples) Pop() interface{} {
	n := len(*h)
	x := (*h)[n-1]
	*h = (*h)[0 : n-1]
	return x
}

// Sample picks n pseudo-randomly chosen input items.  The picking is
// deterministic.
func Sample(n int) Filter {
	return func(arg Arg) {
		// Compute a hash of <index, value> for each item.
		// Keep these hashes in a heap and yield the n items
		// with the largest hashes.
		h := &samples{}
		i := 0
		for s := range arg.In {
			i++
			hash := sha1.Sum([]byte(fmt.Sprintf("%d %s", i, s)))
			heap.Push(h, sample{fmt.Sprintf("%x", hash), s})
			if len(*h) > n {
				heap.Pop(h)
			}
		}
		for len(*h) > 0 {
			arg.Out <- heap.Pop(h).(sample).item
		}
	}
}
